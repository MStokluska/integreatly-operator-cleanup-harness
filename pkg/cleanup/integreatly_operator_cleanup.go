package cleanup

import (
	"context"
	"github.com/integr8ly/integreatly-operator-cleanup-harness/pkg/metadata"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	clientv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"time"
)

const (
	DryRun = "dry-run"
	CleanUpNameSpace = "redhat-rhmi-operator-cleanup-harness"
	ClusterService   = "integreatly-operator-cluster-service"
	ClusterServiceImage = "quay.io/integreatly/cluster-service:v0.4.0"
	Timeout = 35 * time.Minute
	Delay = 30 * time.Second
)

var _ = ginkgo.Describe("Integreatly Operator Cleanup", func() {
	defer ginkgo.GinkgoRecover()
	args := os.Args[1:]

	ginkgo.It("cleanup AWS resources using cluster-service", func() {

		config, err := rest.InClusterConfig()

		if err != nil {
			panic(err)
		}

		// Creates the clientset
		clientV1Set, err := clientv1.NewForConfig(config)
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}
		Expect(err).NotTo(HaveOccurred())

		// get aws values
		secret, err := clientset.CoreV1().Secrets("kube-system").Get(context.TODO(), "aws-creds", metav1.GetOptions{})
		AWS_ACCESS_KEY_ID := string(secret.Data["aws_access_key_id"])
		AWS_SECRET_ACCESS_KEY := string(secret.Data["aws_secret_access_key"])
		logrus.Infof("AWS ACCESS %v", AWS_ACCESS_KEY_ID)
		logrus.Infof("AWS SECRET %v", AWS_SECRET_ACCESS_KEY)

		// get cluster infrastructure
		infrastructure, err := clientV1Set.Infrastructures().Get(context.TODO(), "cluster", metav1.GetOptions{})
		infrastructureName := string(infrastructure.Status.InfrastructureName)
		logrus.Infof("cluster infrastructure ID %v", infrastructureName)

		// configure cluster-service pod args
		container_args := []string{"cleanup", infrastructureName, "--watch"}

		if Contains(args, DryRun) {
			logrus.Info("running cluster-service as dry-run")
			container_args = append(container_args, "--dry-run=true")
		}

		// create a namespace for the cluster service to run in
		CreatedNamespace := v1.Namespace{
			TypeMeta:   metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: CleanUpNameSpace,
			},
		}
		_, err = clientset.CoreV1().Namespaces().Create(context.TODO(), &CreatedNamespace, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}

		// create cluster-service pod
		pod := &v1.Pod{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{Name: ClusterService},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "cluster-service",
						Image: ClusterServiceImage,
						Args:  container_args,
						Env: []v1.EnvVar{
							{
								Name:  "AWS_ACCESS_KEY_ID",
								Value: AWS_ACCESS_KEY_ID,
							}, {
								Name:  "AWS_SECRET_ACCESS_KEY",
								Value: AWS_SECRET_ACCESS_KEY,
							},
						},
					},
				},
				RestartPolicy: "Never",
			},
		}

		logrus.Infof("deploy %v to %v", pod.Name, CleanUpNameSpace)
		_, err = clientset.CoreV1().Pods(CleanUpNameSpace).Create(context.TODO(), pod, metav1.CreateOptions{})

		// watch cluster-service pod for completion
		err = wait.Poll(Timeout, Delay, func() (done bool, err error) {
			pod, err = clientset.CoreV1().Pods(CleanUpNameSpace).Get(context.TODO(), ClusterService, metav1.GetOptions{})
			if err != nil {
				return false, nil
			}

			if pod.Status.Phase == "Succeeded" {
				logrus.Infof("pod %v status is completed", pod.Name)
				return true, nil
			}
			return false, nil
		})

		// add reported value
		if err != nil {
			metadata.Instance.CleanupCompleted = false
		} else {
			metadata.Instance.CleanupCompleted = true
		}

		// remove created namespace
		if !Contains(args, DryRun) {
			logrus.Infof("cleaning up namespace: %v", CleanUpNameSpace)
			err = clientset.CoreV1().Namespaces().Delete(context.TODO(), CleanUpNameSpace, metav1.DeleteOptions{})
			if err != nil {
				metadata.Instance.NameSpaceCleanUp = false
			} else {
				metadata.Instance.NameSpaceCleanUp = true
			}
		} else {
			logrus.Infof("running in dry run mode, namespace %v not cleaned up", CleanUpNameSpace)
			metadata.Instance.NameSpaceCleanUp = false
		}

		Expect(err).NotTo(HaveOccurred())
	})
})

func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

# integreatly-operator-cleanup-harness

This harness sets up the [cluster-service](https://github.com/integr8ly/cluster-service) image with the required aws credentials and cluster infrastructure id.
When running this image the possible arguments are.

- `dry-run` -> set the cluster-service into dry run mode. 
The [cluster-service](https://github.com/integr8ly/cluster-service) will be use in dry run and not remove any resources from aws.
Created namespaces will not be cleaned up. Namespace: `redhat-rhmi-operator-cleanup-harness` 

## Building the image
To run the image on a cluster you must build and push the image to an image repository.
```
podman build -t <path/to/repository>/integreatly-operator-cleanup-harness .
```
```
podman push <path/to/repository>/integreatly-operator-cleanup-harness
```

## Running the image
Run the image with `cluster-admin` permissions on the `<namespace>`.
Set the cluster role bindings.
```
kubectl create clusterrolebinding --user system:serviceaccount:<namespace>:default namespace-cluster-admin --clusterrole cluster-admin
```
Deploy the image into the `<namespace>`.
`dry-run` is optional and will run the cluster-service in a dry run mode.
```
oc run -n <namespace> --restart=Never --image <path/to/repository>/integreatly-operator-cleanup-harness -- integreatly-operator-cleanup-harness [dry-run]
```

## On Cluster

On the cluster, in the `<namespace>` the `integreatly-operator-cleanup-harness` is deployed.
This deployment retrieves the required values and credentials to deploy the [cluster-service](https://github.com/integr8ly/cluster-service).

As the namespace for the cleanup harness may be unknown the `integreatly-operator-cleanup-harnass` creates a namespace to deploy the [cluster-service](https://github.com/integr8ly/cluster-service).
This namespace is `redhat-rhmi-cleanup-harness` which is where the [cluster-service](https://github.com/integr8ly/cluster-service) is deployed.

Under normal running of cleanup harness the `redhat-rhmi-cleanup-harness` namespace will be remove.
With the `dry-run` flag passed this namespace will not be removed.

## Values and Credentials used

- AWS_ACCESS_KEY_ID -> got from the `aws-creds` secret in the `kube-system` namespace.
- AWS_SECRET_ACCESS_KEY -> got from the `aws-creds` secret in the `kube-system` namespace.
- Infrastructure Name -> pass as an argument from the cluster resource `type: infrastructure, name: cluster`.
`oc get infrastructure cluster -o jsonpath='{.status.infrastructureName}{"\n"}'` gives the same value that is pass in. 
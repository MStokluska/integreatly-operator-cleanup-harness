FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13 AS builder

ENV PKG=/go/src/github.com/integr8ly/integreatly-operator-cleanup-harness/
WORKDIR ${PKG}

# compile test binary
COPY . .
RUN make

FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

COPY --from=builder /go/src/github.com/integr8ly/integreatly-operator-cleanup-harness/integreatly-operator-cleanup-harness.test integreatly-operator-cleanup-harness.test

ENTRYPOINT [ "/integreatly-operator-cleanup-harness.test" ]


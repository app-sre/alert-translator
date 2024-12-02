FROM registry.access.redhat.com/ubi8/go-toolset:1.21.13-1.1727869850 as builder
USER 0
WORKDIR /workspace
COPY . .
RUN make test build

FROM registry.access.redhat.com/ubi8-minimal:8.10-1130
COPY --from=builder /workspace/alert-translator  /bin/alert-translator
RUN microdnf update -y && microdnf install -y git && microdnf install -y ca-certificates

ENTRYPOINT  [ "/bin/alert-translator" ]

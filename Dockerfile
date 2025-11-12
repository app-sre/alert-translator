FROM registry.access.redhat.com/ubi9/go-toolset:1.24.4-1753221510 as builder
USER 0
WORKDIR /workspace
COPY . .
RUN make build

FROM builder AS test
RUN make test

FROM registry.access.redhat.com/ubi9-minimal:9.7-1762956380 AS prod
LABEL konflux.additional-tags=1.0.0
COPY --from=builder /workspace/alert-translator  /bin/alert-translator
RUN microdnf update -y && microdnf install -y git && microdnf install -y ca-certificates
ENTRYPOINT  [ "/bin/alert-translator" ]

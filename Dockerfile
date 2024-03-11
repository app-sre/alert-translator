FROM quay.io/app-sre/golang:1.20.6 as builder
WORKDIR /build
COPY . .
RUN make test build

FROM registry.access.redhat.com/ubi8-minimal:8.9
COPY --from=builder /build/alert-translator  /bin/alert-translator
RUN microdnf update -y && microdnf install -y git && microdnf install -y ca-certificates

ENTRYPOINT  [ "/bin/alert-translator" ]

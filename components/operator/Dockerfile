#
# This Dockerfile is used to build dockerregistry-operator image on every pre- and post-submit job
#


# Build the operator binary
FROM europe-docker.pkg.dev/kyma-project/prod/external/library/golang:1.24.4-alpine3.22 AS builder

WORKDIR /workdir

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY components/operator components/operator

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o operator ./components/operator/main.go


# Use distroless as minimal base image to package the operator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --chown=65532:65532 --from=builder /workdir/operator .
COPY --chown=65532:65532  config/docker-registry /module-chart
USER 65532:65532

ENTRYPOINT ["/operator"]

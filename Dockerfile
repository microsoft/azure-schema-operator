# Installer image
FROM cblmariner.azurecr.io/base/core:1.0 AS installer


RUN tdnf install -y dnf unzip

# Install .NET's dependencies into a staging location
RUN mkdir /staging \
  && dnf install -y --releasever=1.0 --installroot /staging \
  prebuilt-ca-certificates \
  glibc \
  krb5 \
  libgcc \
  libstdc++ \
  openssl-libs \
  zlib \
  libunwind \
  icu   

# Clean up staging
RUN rm -rf /staging/etc/dnf \
  && rm -rf /staging/run/* \
  && rm -rf /staging/var/cache/dnf \
  && find /staging/var/log -type f -size +0 -delete

RUN curl -L >sqlpackage.zip https://aka.ms/sqlpackage-linux \
  && unzip sqlpackage.zip -d /staging/sqlpackage \
  && chmod a+x /staging/sqlpackage/sqlpackage \
  && rm sqlpackage.zip


# Build the manager binary
FROM golang:1.18 as builder

ARG delta_kusto_version=0.9.0.105
WORKDIR /workspace

RUN curl -L >delta-kusto-linux.tar.gz https://github.com/microsoft/delta-kusto/releases/download/${delta_kusto_version}/delta-kusto-linux.tar.gz \
  && tar -xzvf delta-kusto-linux.tar.gz -C /tmp \
  && chmod +x /tmp/delta-kusto \
  && rm delta-kusto-linux.tar.gz 

COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# used the dotnet/runtime-deps as base since delta-kusto is a dotnet application
# as it's self contained - no need for the actual dotnet runtime.
# FROM mcr.microsoft.com/dotnet/runtime-deps:6.0.1-cbl-mariner1.0-distroless-amd64
# FROM cblmariner.azurecr.io/distroless/base-debug:1.0

# .NET runtime-deps image
FROM cblmariner.azurecr.io/distroless/minimal:1.0

LABEL org.opencontainers.image.description "Azure-Schema-Operator manages Azure Databases schema on large deployments"
LABEL org.opencontainers.image.url "ghcr.io/microsoft/azure-schema-operator/azureschemaoperator"

ENV \
  # Enable detection of running in a container
  DOTNET_RUNNING_IN_CONTAINER=true

WORKDIR /
COPY --from=installer /staging/ /
COPY --from=builder /tmp/delta-kusto /bin/
COPY --from=builder /workspace/manager .

USER 65532:65532

ENTRYPOINT ["/manager"]

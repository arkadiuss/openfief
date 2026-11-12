# OpenFief's first-party IBKR Flex Query connector plugin (D-49). Built by CI from the repository root
# (scripts/ci-images.sh), for linux/amd64 and linux/arm64: the Go build runs on the build platform and cross-compiles.
# Base images are pinned by digest; Renovate updates tag and digest together.
FROM --platform=$BUILDPLATFORM golang:1.27.4-trixie@sha256:b54b2639de8c1bfe538ec5ffc45f29b1932235ac5f6d63c059853646677112d9 AS build
WORKDIR /src
# The core module and the SDK module build as one Go workspace (go.work). The SDK can't be fetched as a module
# version yet: it is copied in, and the workspace's checksums are in go.work.sum.
COPY go.mod go.sum go.work go.work.sum ./
RUN GOWORK=off go mod download
COPY sdk/go/ sdk/go/
COPY internal/ internal/
COPY plugins/ plugins/
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/plugin ./plugins/ibkr-flex

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
# The plugin's manifest (D-42), signed with the image: written from manifest.json by `task plugins:labels`.
LABEL io.openfief.plugin.manifest='{"capabilities":["connector"],"clusterService":{"description":"an in-cluster HTTPS service to call instead of the IBKR Flex Web Service, at the baseUrl path, its certificate issued for the baseUrl host"},"configSchema":{"additionalProperties":false,"properties":{"baseUrl":{"default":"https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService","description":"The IBKR Flex Web Service, or a mirror of it: an https URL; the install shows its host and the grant approves it","maxLength":2048,"pattern":"^https://[A-Za-z0-9.-]+(:[0-9]{1,5})?(/[^?#\\s]*)?$","type":"string"}},"type":"object"},"crossCheckPrices":true,"egress":["ndcdyn.interactivebrokers.com"],"egressFromConfig":["baseUrl"],"institutions":["ibkr"],"secrets":[{"mountPath":"/etc/openfief/ibkr-flex","role":"flex-token"},{"mountPath":"/etc/openfief/ca-bundle","optional":true,"role":"ca-bundle"}],"statementFiles":true}'
COPY --from=build /out/plugin /plugin
# Numeric, not nonroot:nonroot: the pods run with runAsNonRoot, and the kubelet refuses to start a container
# whose image names its user, since it can't tell that a name isn't root (CreateContainerConfigError).
# 65532 is distroless's nonroot user and group.
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/plugin"]

# OpenFief's first-party IBKR Flex Query connector plugin (D-49). Built by CI from the repository root
# (scripts/ci-images.sh), for linux/amd64 and linux/arm64: the Go build runs on the build platform and cross-compiles.
# Base images are pinned by digest; Renovate updates tag and digest together.
FROM --platform=$BUILDPLATFORM golang:1.27.3-trixie@sha256:c124b9ad3452f46061b5054607d26ea51730dae5372c5e469920063ba2bc6dcc AS build
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
LABEL io.openfief.plugin.manifest='{"capabilities":["connector"],"configSchema":{"additionalProperties":false,"properties":{"accounts":{"additionalProperties":{"minLength":1,"type":"string"},"minProperties":1,"type":"object"}},"required":["accounts"],"type":"object"},"egress":["ndcdyn.interactivebrokers.com"],"secrets":[{"mountPath":"/etc/openfief/ibkr-flex","role":"flex-token"}]}'
COPY --from=build /out/plugin /plugin
# Numeric, not nonroot:nonroot: the pods run with runAsNonRoot, and the kubelet refuses to start a container
# whose image names its user, since it can't tell that a name isn't root (CreateContainerConfigError).
# 65532 is distroless's nonroot user and group.
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/plugin"]

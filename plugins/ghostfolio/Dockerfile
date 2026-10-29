# OpenFief's Ghostfolio view plugin (D-3, D-4, D-64), pushing the record to the owner's Ghostfolio. Built by CI from the repository root
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
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/plugin ./plugins/ghostfolio

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
# The plugin's manifest (D-42), signed with the image: written from manifest.json by `task plugins:labels`.
LABEL io.openfief.plugin.manifest='{"capabilities":["view"],"clusterService":{"description":"a Ghostfolio running in this cluster"},"configSchema":{"additionalProperties":false,"properties":{"interval":{"pattern":"^([0-9]+([.][0-9]+)?(s|m|h))+$","type":"string"},"listings":{"items":{"additionalProperties":false,"properties":{"dataSource":{"enum":["ALPHA_VANTAGE","COINGECKO","EOD_HISTORICAL_DATA","FINANCIAL_MODELING_PREP","GHOSTFOLIO","GOOGLE_SHEETS","RAPID_API","YAHOO"]},"isin":{"pattern":"^[A-Z]{2}[A-Z0-9]{9}[0-9]$","type":"string"},"symbol":{"minLength":1,"type":"string"}},"required":["isin","dataSource","symbol"],"type":"object"},"type":"array"},"url":{"pattern":"^https://[^/?#@]+/?$","type":"string"}},"type":"object"},"egressFromConfig":["url"],"secrets":[{"mountPath":"/etc/openfief/ghostfolio","role":"security-token"}]}'
COPY --from=build /out/plugin /plugin
# Numeric, not nonroot:nonroot: the pods run with runAsNonRoot, and the kubelet refuses to start a container
# whose image names its user, since it can't tell that a name isn't root (CreateContainerConfigError).
# 65532 is distroless's nonroot user and group.
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/plugin"]

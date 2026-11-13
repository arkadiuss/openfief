# OpenFief's first-party OpenAI-compatible model-provider plugin (D-205). Built by CI from the repository root
# (scripts/ci-images.sh), for linux/amd64 and linux/arm64: the Go build runs on the build platform and cross-compiles.
# Base images are pinned by digest; Renovate updates tag and digest together.
FROM --platform=$BUILDPLATFORM golang:1.27.4-trixie@sha256:b54b2639de8c1bfe538ec5ffc45f29b1932235ac5f6d63c059853646677112d9 AS build
WORKDIR /src
# The core module and the SDK module build as one Go workspace (go.work). The SDK can't be fetched as a module
# version yet: it is copied in, and the workspace's checksums are in go.work.sum.
COPY go.mod go.sum go.work go.work.sum ./
RUN GOWORK=off go mod download
COPY sdk/go/ sdk/go/
COPY plugins/ plugins/
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/plugin ./plugins/model-openai

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
# The plugin's manifest (D-42), signed with the image: written from manifest.json by `task plugins:labels`.
LABEL io.openfief.plugin.manifest='{"capabilities":["model-provider"],"clusterService":{"description":"an in-cluster HTTPS service to call instead of an internet provider, such as Ollama or vLLM, at the baseUrl path, its certificate issued for the baseUrl host"},"configSchema":{"additionalProperties":false,"properties":{"baseUrl":{"description":"The OpenAI-compatible API root of the provider, such as https://api.openai.com/v1: an https URL; the install shows its host and the grant approves it","maxLength":2048,"pattern":"^https://[A-Za-z0-9.-]+(:[0-9]{1,5})?(/[^?#\\s]*)?$","type":"string"},"location":{"description":"Where the models of the provider run: on hardware the owner runs, or in the cloud of a third party","enum":["owner_hardware","third_party_cloud"]},"models":{"items":{"additionalProperties":false,"properties":{"contextTokens":{"minimum":0,"type":"integer"},"id":{"maxLength":128,"minLength":1,"type":"string"},"inputPrice":{"description":"Price per million input tokens, a decimal string","pattern":"^[0-9]{1,12}([.][0-9]{1,12})?$","type":"string"},"maxOutputTokens":{"minimum":0,"type":"integer"},"outputPrice":{"description":"Price per million output tokens, a decimal string","pattern":"^[0-9]{1,12}([.][0-9]{1,12})?$","type":"string"},"priceCurrency":{"pattern":"^[A-Z]{3}$","type":"string"},"structuredOutput":{"type":"boolean"},"toolCalling":{"type":"boolean"}},"required":["id"],"type":"object"},"maxItems":64,"minItems":1,"type":"array"},"retention":{"additionalProperties":false,"properties":{"days":{"description":"Days the provider keeps requests and answers; 0 for none","maximum":3650,"minimum":0,"type":"integer"},"mayTrain":{"description":"Whether the provider may train on what it receives","type":"boolean"}},"required":["days","mayTrain"],"type":"object"},"timeout":{"default":"60s","description":"The longest one call to the provider may take, at most 2m","pattern":"^([0-9]+([.][0-9]+)?(s|m))+$","type":"string"}},"required":["baseUrl","location","retention","models"],"type":"object"},"egressFromConfig":["baseUrl"],"secrets":[{"mountPath":"/etc/openfief/model-provider","optional":true,"role":"api-key"},{"mountPath":"/etc/openfief/ca-bundle","optional":true,"role":"ca-bundle"}]}'
COPY --from=build /out/plugin /plugin
# Numeric, not nonroot:nonroot: the pods run with runAsNonRoot, and the kubelet refuses to start a container
# whose image names its user, since it can't tell that a name isn't root (CreateContainerConfigError).
# 65532 is distroless's nonroot user and group.
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/plugin"]

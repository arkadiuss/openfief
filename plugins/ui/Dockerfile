# OpenFief's built-in UI, a view plugin (D-3) serving the owner's browser. Built by CI from the repository root
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
# The design system (design/design.go embeds it): the UI serves its tokens, fonts, marks and Kasztan.
COPY design/ design/
COPY plugins/ plugins/
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/plugin ./plugins/ui

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
# The plugin's manifest (D-42), signed with the image: written from manifest.json by `task plugins:labels`.
LABEL io.openfief.plugin.manifest='{"browserPath":"/","capabilities":["view"],"configSchema":{"additionalProperties":false,"properties":{"askDeadline":{"description":"The longest one Ask box question may take, from 10s to 5m; unset is 2m (D-208)","pattern":"^([0-9]+([.][0-9]+)?(s|m))+$","type":"string"},"askOutputTokens":{"description":"The most output tokens the model may write for one Ask box question, all its calls together; 0 or unset is 800 (D-208)","maximum":4000,"minimum":0,"type":"integer"},"askToolCalls":{"description":"The most tool calls one Ask box question runs in all; 0 or unset is 10 (D-208)","maximum":50,"minimum":0,"type":"integer"},"askToolRounds":{"description":"The most rounds of one Ask box question that may call tools, before one last round without; 0 or unset is 5 (D-208)","maximum":10,"minimum":0,"type":"integer"},"clientEvents":{"description":"Whether the pages report a fixed set of events about themselves (the person check states and error codes, script errors, failed requests, the browser family and version) to the UI log and metrics, never anything typed (D-227); unset is true","type":"boolean"},"turnstileSiteKey":{"description":"The public key of the Cloudflare Turnstile widget made for the host of the Ask box; set together with turnstileVerifyUrl to put the check in front of the model of the Ask box (D-208), unset for none","pattern":"^[0-3]x[A-Za-z0-9_-]{8,64}$","type":"string"},"turnstileVerifyUrl":{"description":"The siteverify URL of Cloudflare, https://challenges.cloudflare.com/turnstile/v0/siteverify; the install shows its host and the grant approves it","maxLength":2048,"pattern":"^https://[A-Za-z0-9.-]+(:[0-9]{1,5})?(/[^?#\\s]*)?$","type":"string"}},"type":"object"},"egressFromConfig":["turnstileVerifyUrl"],"secrets":[{"mountPath":"/etc/openfief/turnstile","optional":true,"role":"turnstile-secret"}]}'
COPY --from=build /out/plugin /plugin
# Numeric, not nonroot:nonroot: the pods run with runAsNonRoot, and the kubelet refuses to start a container
# whose image names its user, since it can't tell that a name isn't root (CreateContainerConfigError).
# 65532 is distroless's nonroot user and group.
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/plugin"]

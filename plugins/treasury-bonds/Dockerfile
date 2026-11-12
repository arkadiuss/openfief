# OpenFief's first-party treasury bond connector plugin (D-9, D-49). Built by CI from the repository root
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
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/plugin ./plugins/treasury-bonds

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
# The plugin's manifest (D-42), signed with the image: written from manifest.json by `task plugins:labels`.
LABEL io.openfief.plugin.manifest='{"bondTerms":true,"capabilities":["connector"],"clusterService":{"description":"an in-cluster HTTPS service to call instead of GUS, NBP, gov.pl and dane.gov.pl, at the gusBaseUrl, nbpBaseUrl, govPlBaseUrl and daneGovPlBaseUrl paths, its certificate issued for their hosts"},"configSchema":{"additionalProperties":false,"properties":{"daneGovPlBaseUrl":{"default":"https://api.dane.gov.pl","description":"The dane.gov.pl API, which lists the same spreadsheet, or a mirror of it: an https URL; the install shows its host and the grant approves it","maxLength":2048,"pattern":"^https://[A-Za-z0-9.-]+(:[0-9]{1,5})?(/[^?#\\s]*)?$","type":"string"},"govPlBaseUrl":{"default":"https://www.gov.pl","description":"gov.pl, whose Ministry of Finance page links the retail bond spreadsheet, or a mirror of it: an https URL; the install shows its host and the grant approves it","maxLength":2048,"pattern":"^https://[A-Za-z0-9.-]+(:[0-9]{1,5})?(/[^?#\\s]*)?$","type":"string"},"gusBaseUrl":{"default":"https://api-sdp.stat.gov.pl","description":"The GUS SDP API, or a mirror of it: an https URL; the install shows its host and the grant approves it","maxLength":2048,"pattern":"^https://[A-Za-z0-9.-]+(:[0-9]{1,5})?(/[^?#\\s]*)?$","type":"string"},"nbpBaseUrl":{"default":"https://static.nbp.pl","description":"The NBP archive, or a mirror of it: an https URL; the install shows its host and the grant approves it","maxLength":2048,"pattern":"^https://[A-Za-z0-9.-]+(:[0-9]{1,5})?(/[^?#\\s]*)?$","type":"string"},"start":{"maxLength":10,"minLength":10,"pattern":"^[0-9]{4}-[0-9]{2}-[0-9]{2}","type":"string"}},"type":"object"},"egress":["api-sdp.stat.gov.pl","static.nbp.pl","www.gov.pl","api.dane.gov.pl"],"egressFromConfig":["gusBaseUrl","nbpBaseUrl","govPlBaseUrl","daneGovPlBaseUrl"],"indexReadings":true,"institutions":["treasury-bonds"],"secrets":[{"mountPath":"/etc/openfief/ca-bundle","optional":true,"role":"ca-bundle"}]}'
COPY --from=build /out/plugin /plugin
# Numeric, not nonroot:nonroot: the pods run with runAsNonRoot, and the kubelet refuses to start a container
# whose image names its user, since it can't tell that a name isn't root (CreateContainerConfigError).
# 65532 is distroless's nonroot user and group.
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/plugin"]

# The end-to-end suite's synthetic connector (main.go), built only by scripts/e2e-images.sh from the repository root
# and never released. MANIFEST is the manifest the image carries as its signed label (D-42): manifest.json, or
# manifest-egress.json for the image the suite upgrades to; the suite declares the same one in each Plugin.
# Base images are pinned by digest, as plugins/nbp-fx/Dockerfile pins them; Renovate updates tag and digest together.
FROM golang:1.27.4-trixie@sha256:b54b2639de8c1bfe538ec5ffc45f29b1932235ac5f6d63c059853646677112d9 AS build
WORKDIR /src
# The core module and the SDK module build as one Go workspace (go.work).
COPY go.mod go.sum go.work go.work.sum ./
RUN GOWORK=off go mod download
COPY sdk/go/ sdk/go/
COPY test/e2e/connector/ test/e2e/connector/
RUN CGO_ENABLED=0 go build -tags e2e -trimpath -ldflags='-s -w' -o /out/plugin ./test/e2e/connector

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
ARG MANIFEST
LABEL io.openfief.plugin.manifest="${MANIFEST}"
COPY --from=build /out/plugin /plugin
# Numeric, as the first-party plugins run: the pods run with runAsNonRoot.
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/plugin"]

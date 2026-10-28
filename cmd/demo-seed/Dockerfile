# OpenFief demo data loader: loads the synthetic demo dataset into a demo database and exits.
# Built by CI from the repository root (scripts/ci-images.sh), for linux/amd64 and linux/arm64:
# the Go build runs on the build platform and cross-compiles. It imports nothing from sdk/go, so
# go.mod alone builds it (a binary that imports sdk/go would also need go.work and sdk/go copied in).
# Base images are pinned by digest; Renovate updates tag and digest together.
FROM --platform=$BUILDPLATFORM golang:1.27.3-trixie@sha256:c124b9ad3452f46061b5054607d26ea51730dae5372c5e469920063ba2bc6dcc AS build
WORKDIR /src
# The core module and the SDK module build as one Go workspace (go.work). The repository is private, so
# the SDK can't be fetched as a module version: it is copied in, and the workspace's checksums are in go.work.sum.
COPY go.mod go.sum go.work go.work.sum ./
RUN GOWORK=off go mod download
COPY sdk/go/ sdk/go/
COPY cmd/ cmd/
COPY internal/ internal/
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/demo-seed ./cmd/demo-seed

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/demo-seed /demo-seed
USER nonroot:nonroot
ENTRYPOINT ["/demo-seed"]

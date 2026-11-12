# The end-to-end suite's institution fixture server (main.go), built only by scripts/e2e-images.sh from the repository
# root and never released. It serves synthetic IBKR Flex, EODHD and NBP answers over HTTPS with a certificate the suite
# makes for each run. Base images are pinned by digest, as plugins/nbp-fx/Dockerfile pins them; Renovate updates tag
# and digest together.
FROM golang:1.27.4-trixie@sha256:b54b2639de8c1bfe538ec5ffc45f29b1932235ac5f6d63c059853646677112d9 AS build
WORKDIR /src
# The core module and the SDK module build as one Go workspace (go.work).
COPY go.mod go.sum go.work go.work.sum ./
RUN GOWORK=off go mod download
COPY sdk/go/ sdk/go/
# The fixture's handlers use internal/kernel's decimal types, as the plugins' builds copy internal/.
COPY internal/ internal/
# The model portfolio's synthetic index series are shared with the suite's hand calculation (test/e2e/modelseries).
COPY test/e2e/modelseries/ test/e2e/modelseries/
COPY test/e2e/fixture/ test/e2e/fixture/
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/fixture ./test/e2e/fixture

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/fixture /fixture
# Numeric, as the plugins run: the namespace enforces runAsNonRoot.
USER 65532:65532
EXPOSE 8443
ENTRYPOINT ["/fixture"]

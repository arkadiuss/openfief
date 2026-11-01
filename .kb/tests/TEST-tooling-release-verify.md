---
title: release-verify.sh verifies the CLI and the install's images as the install guide does, and the guide names the identities the release signs with
status: active
sourceFile: scripts/tests/release-verify.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-01T14:57:59.689Z'
id: TEST-tooling-release-verify
type: test
---
Runs scripts/release-verify.sh with a fake cosign whose verify-blob accepts a bundle only when it names the identity and the file's sha256 it was given with GitHub's issuer, and whose verify accepts only listed refs. A synthetic CLI directory signed as release.yml on main verifies, every file checked with that identity; one signed at another ref verifies only at that ref, one signed by images.yml fails, and a changed binary, a missing bundle or a missing SHA256SUMS fail. An install file's images verify by digest as images.yml on main, each once; an unsigned digest, a signature for another digest and an image by tag fail; a dry run's all-zeros placeholders verify only at the given tag. Then it reads docs/guides/install.md and checks that the guide's --certificate-identity values are exactly the two identities the script derives for arkadiuss/openfief, that its only issuer is GitHub Actions', that internal/plugins/trust.go's OpenFiefCISubject is the images identity, and that release.yml signs the CLI, images.yml signs the images and release.yml publishes only from main. A false pass would be a guide whose identity no release ever signs with, which this comparison catches.
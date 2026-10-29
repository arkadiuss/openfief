---
title: The release adds each first-party plugin to the catalog with the digest it signed
status: active
sourceFile: scripts/tests/release-catalog.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:53.856Z'
id: TEST-tooling-release-catalog
type: test
---
Runs scripts/release-catalog.sh in a throwaway directory with two plugins and synthetic digests: both enter the catalog with the release version, the lower-cased image at its digest and the publisher openfief with OpenFief's CI identity; a plugin without a digest, a digest that is not sha256 and a name and version already listed fail; a next version keeps the earlier entries. Signature verification (VERIFY=1) needs cosign and runs only in CI.
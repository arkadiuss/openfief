---
title: The install guide names exactly the identities the release signs with
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:26.139Z'
id: SCEN-docs-install-guide-identities
type: scenario
---
Given the install guide's cosign commands, when the scripts' test compares them with the identities release-verify.sh derives for arkadiuss/openfief, then the guide verifies the CLI as release.yml on main and the images as images.yml on main, both with GitHub Actions' issuer, names no other identity or issuer, and the operator trusts the same images identity; renaming a workflow or moving the signing breaks the test.
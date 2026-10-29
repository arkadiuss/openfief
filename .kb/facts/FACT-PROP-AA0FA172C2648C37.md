---
title: Signed Cli Every Release = true
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.release_cli
property_key: signed_cli_every_release
canonical_key: mcp-kibi-model-requirement:tooling.release_cli:signed_cli_every_release:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4E9C525AC6B06AE7
claim_text: Every release must build the openfief CLI as static binaries for Linux and macOS on amd64 and arm64 with an SBOM and SHA256SUMS, sign each file with a keyless cosign bundle, and attach them to its GitHub release.
origin:
  kind: agent
  recorded_at: '2026-10-29T18:02:14.786Z'
id: FACT-PROP-AA0FA172C2648C37
type: fact
---

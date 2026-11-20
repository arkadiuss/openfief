---
title: Tooling Release Cli
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: tooling.release_cli
canonical_key: tooling.release_cli
claim_key: CLAIM-4E9C525AC6B06AE7
claim_text: Every release must build the openfief CLI as static binaries for Linux and macOS on amd64 and arm64 with an SBOM and SHA256SUMS, sign each file with a keyless cosign bundle, and attach them to its GitHub release.
origin:
  kind: agent
  recorded_at: '2026-10-29T18:02:12.524Z'
id: FACT-SUBJECT-5EDF36412545D9BB
type: fact
---

Definition not stated: the subject tooling.release_cli is named by requirement REQ-tooling-release-cli-signed, which links to the subject fact FACT-SUBJECT-5EDF36412545D9BB through constrains. The people who own that source can say what it is in the product and what it is not.

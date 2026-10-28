---
title: The operator must verify signatures against a Sigstore trust root embedded in its binary, without network access to Sigstore services, and the embedded root must be refreshed only through a TUF-verified update procedure.
status: open
text_ref: docs/DESIGN.md#D-56
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-56
  - lane:strict
semantic_text: The operator must verify signatures against a Sigstore trust root embedded in its binary, without network access to Sigstore services, and the embedded root must be refreshed only through a TUF-verified update procedure.
logic_claims:
  - CLAIM-73CEC9C6009BDDAF
semantic_clauses:
  - The operator must verify signatures against a Sigstore trust root embedded in its binary, without network access to Sigstore services, and the embedded root must be refreshed only through a TUF-verified update procedure.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 409da0a7883155f20da229fa81bd11c4ee708bbdf25b84b6dfa7efaa4b27fba1
semantic_inventory:
  - claim_key: CLAIM-73CEC9C6009BDDAF
    claim_text: The operator must verify signatures against a Sigstore trust root embedded in its binary, without network access to Sigstore services, and the embedded root must be refreshed only through a TUF-verified update procedure
    role: normative
    status: modeled
    span:
      start: 0
      end: 219
origin:
  kind: agent
  recorded_at: '2026-10-28T09:17:22.309Z'
id: REQ-plugins-bundled-trust-root
type: req
---
The operator must verify signatures against a Sigstore trust root embedded in its binary, without network access to Sigstore services, and the embedded root must be refreshed only through a TUF-verified update procedure.

## Context

D-56 requires a bundled trust root so verification works offline. The public-good trusted_root.json is embedded; task trustroot:update fetches it with sigstore-go's TUF client, starting from the TUF root sigstore-go embeds, checks that it parses, and writes it for review and commit. Refines REQ-plugins-operator-verifies-signatures. Modeled by Claude Code for the M0 signature-verification task; the author has not reviewed it yet.

## Source

> It uses a bundled trust root, so verification works offline.

docs/DESIGN.md D-56

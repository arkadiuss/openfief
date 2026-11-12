---
title: Downloads CI makes of its own, tool installs, test assets, release manifests and the cosign installer included, must be retried with backoff when the server answers 429 or 5xx or the connection fails, must never be retried on any other 4xx answer, and must keep their checksum checks as strict as before.
status: open
text_ref: docs/dev/sdlc.md#tooling
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-tooling
  - lane:strict
semantic_text: Downloads CI makes of its own, tool installs, test assets, release manifests and the cosign installer included, must be retried with backoff when the server answers 429 or 5xx or the connection fails, must never be retried on any other 4xx answer, and must keep their checksum checks as strict as before.
logic_claims:
  - CLAIM-A4193AACD985DF55
semantic_clauses:
  - Downloads CI makes of its own, tool installs, test assets, release manifests and the cosign installer included, must be retried with backoff when the server answers 429 or 5xx or the connection fails, must never be retried on any other 4xx answer, and must keep their checksum checks as strict as before.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a0f9bfaa53c9c230ee8a769acb9c5d1943337b55165a509bb844fe7db9899b2c
semantic_inventory:
  - claim_key: CLAIM-A4193AACD985DF55
    claim_text: Downloads CI makes of its own, tool installs, test assets, release manifests and the cosign installer included, must be retried with backoff when the server answers 429 or 5xx or the connection fails, must never be retried on any other 4xx answer, and must keep their checksum checks as strict as before
    role: normative
    status: modeled
    span:
      start: 0
      end: 303
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:58.135Z'
id: REQ-tooling-download-retry
type: req
---
Downloads CI makes of its own, tool installs, test assets, release manifests and the cosign installer included, must be retried with backoff when the server answers 429 or 5xx or the connection fails, must never be retried on any other 4xx answer, and must keep their checksum checks as strict as before.

## Context

The envtest download failed a CI run with curl's exit 22 on a 504, and curl's own retry waited a few seconds in all; a retry that also hid a 404 or a failed checksum would hide a real fault, so only failures a later try can fix are retried. The author approved the fast-gate plan (the pre-commit hook as a fast tier, the full gate in CI, retried downloads); implemented by Claude Code on the build/fast-gate branch, and the author has not reviewed this wording yet. 

## Source

> It never retries anything else: a 4xx answer other than 429, a certificate problem, a failed checksum or a compile error is final at once.

docs/dev/sdlc.md#tooling

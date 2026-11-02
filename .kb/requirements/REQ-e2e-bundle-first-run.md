---
title: 'An end-to-end test must follow the owner''s first run on an install from the release bundle: the openfief CLI built from the commit installs the UI plugin and enrolls the first passkey.'
status: open
text_ref: docs/dev/testing.md#end-to-end-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-testing-md-end-to-end-tests
  - lane:strict
semantic_text: 'An end-to-end test must follow the owner''s first run on an install from the release bundle: the openfief CLI built from the commit installs the UI plugin and enrolls the first passkey.'
logic_claims:
  - CLAIM-61A1B3D12851EC11
semantic_clauses:
  - 'An end-to-end test must follow the owner''s first run on an install from the release bundle: the openfief CLI built from the commit installs the UI plugin and enrolls the first passkey.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 42f5fb33b6a2e70b0afb7d7d45871cba62de9502c2b4d002e2a747ef31760dbf
semantic_inventory:
  - claim_key: CLAIM-61A1B3D12851EC11
    claim_text: 'An end-to-end test must follow the owner''s first run on an install from the release bundle: the openfief CLI built from the commit installs the UI plugin and enrolls the first passkey'
    role: normative
    status: modeled
    span:
      start: 0
      end: 183
origin:
  kind: agent
  recorded_at: '2026-11-01T21:52:12.392Z'
id: REQ-e2e-bundle-first-run
type: req
---
An end-to-end test must follow the owner's first run on an install from the release bundle: the openfief CLI built from the commit installs the UI plugin and enrolls the first passkey.

## Context

The author's testing rules ask for it, so that the first alpha's install path is tested as an owner follows it (docs/guides/install.md, steps 6 and 7). Only downloading the CLI from a release stays outside the suite.

## Source

> An end-to-end test follows the owner's first run: the `openfief` CLI built from the commit installs the UI plugin and enrolls the first passkey. Only downloading the CLI from a release stays outside the suite.

docs/dev/testing.md#end-to-end-tests

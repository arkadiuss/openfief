---
title: The core holds a thin contract and the model router, which gives each role a primary model and ordered fallbacks as provider/model references.
status: open
text_ref: design-decisions:D-205#3
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-205#3
  recorded_at: '2026-11-13T11:29:08.589Z'
semantic_text: The core holds a thin contract and the model router, which gives each role a primary model and ordered fallbacks as provider/model references.
semantic_source_field: semantic_text
semantic_source_hash: 8f7c1fe0e1cab47ca3a90f9e88694ff7c361fcc59fb0422f5e2f4920f3bc052e
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-018F388937521E4A
    claim_text: The core holds a thin contract and the model router, which gives each role a primary model and ordered fallbacks as provider/model references
    role: descriptive
    status: missing
    span:
      start: 0
      end: 141
logic_claims:
  - CLAIM-018F388937521E4A
semantic_clauses:
  - The core holds a thin contract and the model router, which gives each role a primary model and ordered fallbacks as provider/model references
required_evidence: asserted
id: REQ-models-router-roles-fallbacks
type: req
---
The core holds a thin contract and the model router, which gives each role a primary model and ordered fallbacks as provider/model references.

## Context

A model that is unavailable passes the request to the route's next model; a model that rejects the request ends the call.

## Source

> The core holds a thin contract and the model router, which gives each role, such as the Ask box (D-203) and the maintainer (D-204), a primary model and ordered fallbacks as `provider/model` references.

Source: OpenFief author decisions (DESIGN.md section 10) - D-205#3
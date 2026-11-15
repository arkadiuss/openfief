---
title: 'The demo''s model spend has a hard limit at the provider: a prepaid OpenRouter key with a monthly credit limit and no automatic top-up.'
status: open
firmness: firm
semantic_text: 'The demo''s model spend has a hard limit at the provider: a prepaid OpenRouter key with a monthly credit limit and no automatic top-up.'
origin:
  kind: import
  ref: design-decisions:D-208#1
  recorded_at: '2026-11-14T20:37:35.005Z'
text_ref: design-decisions:D-208#1
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a7f9b45296d84a33f7fb702cff3414cf9b99f8888e2ba36a3c8a800e1b6349fe
semantic_inventory:
  - claim_key: CLAIM-13050D9DB7062DBE
    claim_text: 'The demo''s model spend has a hard limit at the provider: a prepaid OpenRouter key with a monthly credit limit and no automatic top-up'
    role: descriptive
    status: missing
    span:
      start: 0
      end: 133
logic_claims:
  - CLAIM-13050D9DB7062DBE
id: REQ-decision-d208-1-provider-hard-limit
type: req
---
The demo's model spend has a hard limit at the provider: a prepaid OpenRouter key with a monthly credit limit and no automatic top-up.

## Context

Author decision D-208 (the demo's spend guards), imported from docs/DESIGN.md; it was not imported before. It is the owner's, set at OpenRouter and not in code; docs/dev/demo.md says so and the demo-model-secret task puts the key in the cluster.

## Source

> docs/DESIGN.md, D-208#1.
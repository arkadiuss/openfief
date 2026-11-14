---
title: The model-portfolio maintainer has no trading tools.
status: open
text_ref: design-decisions:D-204#1
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-204#1
  recorded_at: '2026-11-13T21:07:36.532Z'
semantic_text: The model-portfolio maintainer has no trading tools.
semantic_source_field: semantic_text
semantic_source_hash: 8c2497fbef5e507b3bc30ecd4a664de8b4a2322b7f9016f07bbeae4e86b5226c
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-3D80C81D30BD6F0C
    claim_text: The model-portfolio maintainer has no trading tools
    role: descriptive
    status: missing
    span:
      start: 0
      end: 51
logic_claims:
  - CLAIM-3D80C81D30BD6F0C
semantic_clauses:
  - The model-portfolio maintainer has no trading tools
required_evidence: asserted
id: REQ-decision-d204-1-no-trading-tools
type: req
---
The model-portfolio maintainer has no trading tools.

## Context

Author decision D-204 (the read-only models slice, D-202). The maintainer sends its model figures in a prompt and offers it no tool at all.

## Source

> **[D-204]** The model-portfolio maintainer has no trading tools. OpenFief computes the weights and drift; the model writes a note about them weekly and whenever the owner asks, and may propose target changes, which land in the Inbox under decisions with the figures behind them for the owner to apply or dismiss.

Source: OpenFief author decisions (DESIGN.md section 10) - D-204#1
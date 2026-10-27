---
title: Plugins must authenticate to the core with projected ServiceAccount tokens with audience openfief-core.
status: open
text_ref: design-decisions:D-27#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-27-1
  - lane:strict
semantic_text: Plugins must authenticate to the core with projected ServiceAccount tokens with audience openfief-core.
logic_claims:
  - CLAIM-34471B9A284A2830
semantic_clauses:
  - Plugins must authenticate to the core with projected ServiceAccount tokens with audience openfief-core.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4a53502eafb42857a34467c29607f8b2f1bf1e42cfcb3db8b65c0024411d3a84
semantic_inventory:
  - claim_key: CLAIM-34471B9A284A2830
    claim_text: Plugins must authenticate to the core with projected ServiceAccount tokens with audience openfief-core
    role: normative
    status: modeled
    span:
      start: 0
      end: 102
origin:
  kind: import
  ref: design-decisions:D-27#1
  recorded_at: '2026-10-27T20:12:44.419Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-C6544ADAF8F5AE65
type: req
---
Plugins must authenticate to the core with projected ServiceAccount tokens with audience openfief-core.

## Source

> Plugins authenticate to the core with projected ServiceAccount tokens (audience `openfief-core`). Scripts get separate, scoped, expiring, read-only personal tokens.

Source: OpenFief author decisions (DESIGN.md section 10) - D-27#1

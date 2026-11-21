---
title: A market quote of an instrument the owner holds must come from the source its valuation uses, so that it equals that source's latest price, and from no other source, whichever the owner ranks first.
status: open
firmness: default
required_evidence: unverified
semantic_text: A market quote of an instrument the owner holds must come from the source its valuation uses, so that it equals that source's latest price, and from no other source, whichever the owner ranks first.
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
text_ref: design-decisions:D-226#1
rationale: D-94 gives every instrument one source of its values and D-226 lets the model fetch any instrument's price; two prices of one instrument from two sources would put different numbers in one answer. The design session ruled that a held instrument's quote comes from its valuation source; the mechanism, a lookup through the sources of the valuation's price series, is the implementer's.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d890e6f5bc7d8c1d4760aed526c2b65dcf3d0bf181d769f66c6bd6cc163ffb04
semantic_inventory:
  - claim_key: CLAIM-432A63A17D93ABA1
    claim_text: A market quote of an instrument the owner holds must come from the source its valuation uses, so that it equals that source's latest price, and from no other source, whichever the owner ranks first
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 197
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - ingest.held_instrument_price
logic_claims:
  - CLAIM-432A63A17D93ABA1
id: REQ-ingest-market-quote-held-source
type: req
---
A market quote of an instrument the owner holds must come from the source its valuation uses, so that it equals that source's latest price, and from no other source, whichever the owner ranks first.

## Context

D-94 gives every instrument one source of its values and D-226 lets the model fetch any instrument's price; two prices of one instrument from two sources would put different numbers in one answer. The design session ruled that a held instrument's quote comes from its valuation source; the mechanism, a lookup through the sources of the valuation's price series, is the implementer's.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1

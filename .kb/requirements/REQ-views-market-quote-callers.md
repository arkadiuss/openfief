---
title: The read API must answer a market quote to the owner, a script token and a viewer and must refuse it to a view plugin and a demo session as permission denied without asking the market-data connector, and GetSettings must say quotes are not available to them.
status: open
firmness: default
required_evidence: unverified
semantic_text: The read API must answer a market quote to the owner, a script token and a viewer and must refuse it to a view plugin and a demo session as permission denied without asking the market-data connector, and GetSettings must say quotes are not available to them.
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
text_ref: design-decisions:D-226#1
rationale: D-137 lists quotes among what read access covers; the design session ruled the owner, script tokens and viewers may read them, view plugins only where their manifest asks for them (Ghostfolio's does not, and no manifest can ask yet) and demo sessions never, since the demo answers from its synthetic data alone.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8d8f6c3b0cae902c60e17612a1895f50b9e8421ad7eea1e38189a28d75dc4fc3
semantic_inventory:
  - claim_key: CLAIM-41694AE30481FB2B
    claim_text: The read API must answer a market quote to the owner, a script token and a viewer and must refuse it to a view plugin and a demo session as permission denied without asking the market-data connector, and GetSettings must say quotes are not available to them
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 257
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - views.market_quote_callers
logic_claims:
  - CLAIM-41694AE30481FB2B
id: REQ-views-market-quote-callers
type: req
---
The read API must answer a market quote to the owner, a script token and a viewer and must refuse it to a view plugin and a demo session as permission denied without asking the market-data connector, and GetSettings must say quotes are not available to them.

## Context

D-137 lists quotes among what read access covers; the design session ruled the owner, script tokens and viewers may read them, view plugins only where their manifest asks for them (Ghostfolio's does not, and no manifest can ask yet) and demo sessions never, since the demo answers from its synthetic data alone.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1

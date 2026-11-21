---
title: A market quote must be asked only of a running market-data connector whose manifest declares it and whose grant approves it, taking the first of the owner's ranked sources of prices that knows the instrument, passing over a source that has no listing or takes no quotes and never over one that is down, and the core must make no request of its own to a provider.
status: open
firmness: default
required_evidence: unverified
semantic_text: A market quote must be asked only of a running market-data connector whose manifest declares it and whose grant approves it, taking the first of the owner's ranked sources of prices that knows the instrument, passing over a source that has no listing or takes no quotes and never over one that is down, and the core must make no request of its own to a provider.
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
text_ref: design-decisions:D-226#1
semantic_inventory:
  - claim_key: CLAIM-5E5F1064235F9BF7
    claim_text: A market quote must be asked only of a running market-data connector whose manifest declares it and whose grant approves it, taking the first of the owner's ranked sources of prices that knows the instrument, passing over a source that has no listing or takes no quotes and never over one that is down, and the core must make no request of its own to a provider
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 361
    reason: It states how a market quote is looked up and answered; no predicate of the ontology states it yet.
    subjects:
      - ingest.market_quote_source
logic_claims:
  - CLAIM-5E5F1064235F9BF7
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a620d514b2fb7646aa4548c571bd79074f3870dd18435184de09b0106d1e0e63
rationale: D-226 says the prices come through the owner's own market-data connector and D-94 ranks the sources of prices; D-123 lets only declared and approved connectors learn what the owner asks. The author decided those on 2026-11-21 and earlier, and the design session ruled that a source with no such instrument is passed over and a source that is down is not; how they combine for a lookup is the implementer's, for the author to review.
id: REQ-ingest-market-quote-source
type: req
---
A market quote must be asked only of a running market-data connector whose manifest declares it and whose grant approves it, taking the first of the owner's ranked sources of prices that knows the instrument, passing over a source that has no listing or takes no quotes and never over one that is down, and the core must make no request of its own to a provider.

## Context

D-226 says the prices come through the owner's own market-data connector and D-94 ranks the sources of prices; D-123 lets only declared and approved connectors learn what the owner asks. The author decided those on 2026-11-21 and earlier, and the design session ruled that a source with no such instrument is passed over and a source that is down is not; how they combine for a lookup is the implementer's, for the author to review.

## Source

> On the owner's install, the Ask box's model may fetch prices for any instrument through the owner's own market-data connector. They are T2 data (D-206), so only a model allowed for T2 receives them, under the owner's key and its terms.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-226#1

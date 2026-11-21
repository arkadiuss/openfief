---
title: 'A market quote''s query is at most 32 letters and digits with single dots or dashes between them, checked by the core before it reaches a connector, by the Ask box''s market price tool and again by the EODHD connector: a query that starts or ends with a dot or dash, holds two in a row, a slash or an underscore is refused as an invalid argument with a fixed message that quotes nothing of it.'
status: open
firmness: default
semantic_text: 'A market quote''s query is at most 32 letters and digits with single dots or dashes between them, checked by the core before it reaches a connector, by the Ask box''s market price tool and again by the EODHD connector: a query that starts or ends with a dot or dash, holds two in a row, a slash or an underscore is refused as an invalid argument with a fixed message that quotes nothing of it.'
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #5: validate the quote query to a strict charset in the core and mirror it in the eodhd plugin (the coordinator''s follow-up, 2026-11-21)'
  recorded_at: '2026-11-21T18:46:05.334Z'
rationale: A query of dots alone became a path segment such as .. in the provider's URL, sent with the owner's API key; the strict shape keeps every query a single plain path segment.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3c5faee5552dea0ddbc3f2a77b3bc3872d5e3efe4495e6333073bfd1eea09238
semantic_inventory:
  - claim_key: CLAIM-26D1C7A55D598500
    claim_text: 'A market quote''s query is at most 32 letters and digits with single dots or dashes between them, checked by the core before it reaches a connector, by the Ask box''s market price tool and again by the EODHD connector: a query that starts or ends with a dot or dash, holds two in a row, a slash or an underscore is refused as an invalid argument with a fixed message that quotes nothing of it'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 390
    subjects:
      - quote.query_charset
    reason: It states the shape of a text field and where it is checked; no predicate of the ontology states a string's pattern yet.
logic_claims:
  - CLAIM-26D1C7A55D598500
id: REQ-quote-query-strict-charset
type: req
---
A market quote's query is at most 32 letters and digits with single dots or dashes between them, checked by the core before it reaches a connector, by the Ask box's market price tool and again by the EODHD connector: a query that starts or ends with a dot or dash, holds two in a row, a slash or an underscore is refused as an invalid argument with a fixed message that quotes nothing of it.

## Context

A query of dots alone became a path segment such as .. in the provider's URL, sent with the owner's API key; the strict shape keeps every query a single plain path segment.

## Source

> Security review of the new surfaces before the alpha.2 cut: GetMarketQuote and connector egress (2026-11-21); the coordinator asked to fix known limit #5.

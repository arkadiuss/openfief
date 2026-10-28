---
title: The IBKR Flex parser must refuse a statement over its size, nesting depth, element count, attribute count or attribute length limits, any DOCTYPE, DTD or entity, and non-UTF-8 input, and must read amounts only as exact decimals in plain notation.
status: open
text_ref: docs/DESIGN.md#3.5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-3-5
  - lane:strict
semantic_text: The IBKR Flex parser must refuse a statement over its size, nesting depth, element count, attribute count or attribute length limits, any DOCTYPE, DTD or entity, and non-UTF-8 input, and must read amounts only as exact decimals in plain notation.
logic_claims:
  - CLAIM-EDF43E2B11A65C0F
semantic_clauses:
  - The IBKR Flex parser must refuse a statement over its size, nesting depth, element count, attribute count or attribute length limits, any DOCTYPE, DTD or entity, and non-UTF-8 input, and must read amounts only as exact decimals in plain notation.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cf180bde1760c214e67de431f439a0d6b6beaf98820de9bfd32997cfcbc12bcb
semantic_inventory:
  - claim_key: CLAIM-EDF43E2B11A65C0F
    claim_text: The IBKR Flex parser must refuse a statement over its size, nesting depth, element count, attribute count or attribute length limits, any DOCTYPE, DTD or entity, and non-UTF-8 input, and must read amounts only as exact decimals in plain notation
    role: normative
    status: modeled
    span:
      start: 0
      end: 245
origin:
  kind: agent
  recorded_at: '2026-10-28T06:24:08.994Z'
id: REQ-ibkrflex-hostile-input
type: req
---
The IBKR Flex parser must refuse a statement over its size, nesting depth, element count, attribute count or attribute length limits, any DOCTYPE, DTD or entity, and non-UTF-8 input, and must read amounts only as exact decimals in plain notation.

## Context

DESIGN section 3.5 lists malicious statement files as a threat to the tracking and imports use case. Go's encoding/xml never fetches external entities, but limits and an outright refusal of directives keep a crafted file from exhausting memory or smuggling entities, and plain-notation decimals stop a short exponent from expanding into a huge number. Amounts are never parsed through float64 (AGENTS.md). Modeled by Claude Code for the M0 IBKR Flex connector task; the author has not reviewed it yet. The Flex schema is IBKR's public Flex Query reference, not yet verified against a real account (OQ-22).

## Source

> Threats: theft of the data from the host or its backups; leaked read-only credentials, which still expose the full account history; malicious statement files; wrong valuations leading to wrong decisions.

docs/DESIGN.md section 3.5

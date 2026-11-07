---
title: Outbound traffic must go only to configured destinations, such as IBKR's Flex service, NBP, GUS, the market-data vendor and the user's Ghostfolio, as verified from the egress proxy's logs.
status: open
text_ref: design-decisions:D-200#16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-16
  - lane:strict
semantic_text: Outbound traffic must go only to configured destinations, such as IBKR's Flex service, NBP, GUS, the market-data vendor and the user's Ghostfolio, as verified from the egress proxy's logs.
logic_claims:
  - CLAIM-52E8C5F1BF11D050
semantic_clauses:
  - Outbound traffic must go only to configured destinations, such as IBKR's Flex service, NBP, GUS, the market-data vendor and the user's Ghostfolio, as verified from the egress proxy's logs.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0c452249a3ebb1782827f6abb87e8db9e557e608ff4966849d8cd094939ea62f
semantic_inventory:
  - claim_key: CLAIM-52E8C5F1BF11D050
    claim_text: Outbound traffic must go only to configured destinations, such as IBKR's Flex service, NBP, GUS, the market-data vendor and the user's Ghostfolio, as verified from the egress proxy's logs
    role: normative
    status: modeled
    span:
      start: 0
      end: 187
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#16
  recorded_at: '2026-11-07T03:32:39.505Z'
id: REQ-m0-exit-16-egress
type: req
---
Outbound traffic must go only to configured destinations, such as IBKR's Flex service, NBP, GUS, the market-data vendor and the user's Ghostfolio, as verified from the egress proxy's logs.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-16.

## Source

> **Egress.** Outbound traffic goes only to configured destinations (IBKR's Flex service, NBP, GUS, the market-data vendor, the user's Ghostfolio), verified from the egress proxy's logs.

design-decisions:D-200#16 (docs/DESIGN.md, D-200, criterion 16)

## Why this changed

An earlier agent modeled criterion 16 as the default-firmness requirement REQ-e2e-egress-configured-destinations before D-200's criteria were imported as firm requirements. D-200 is the author's decision, so this firm requirement, stated from criterion 16 itself, replaces it, and the end-to-end scenarios that verified REQ-e2e-egress-configured-destinations now specify this requirement too. Its detail that the proxy's log holds each request's host, decision and instance and nothing of its path, query, headers, body or token stays as the default-firmness refinement REQ-e2e-egress-log-holds-no-request-contents.

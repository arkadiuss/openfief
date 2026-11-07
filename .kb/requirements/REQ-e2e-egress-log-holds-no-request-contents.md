---
title: On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance, and nothing of its path, query, headers, body or token.
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance, and nothing of its path, query, headers, body or token.
logic_claims:
  - CLAIM-66A05045510D7407
semantic_clauses:
  - On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance, and nothing of its path, query, headers, body or token.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4db54e9d5b3285e34174a953adc8190e50d1cdd9ec6b918c771a147f00cb71d8
semantic_inventory:
  - claim_key: CLAIM-66A05045510D7407
    claim_text: On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance, and nothing of its path, query, headers, body or token
    role: normative
    status: modeled
    span:
      start: 0
      end: 182
origin:
  kind: agent
  recorded_at: '2026-11-07T03:31:13.845Z'
id: REQ-e2e-egress-log-holds-no-request-contents
type: req
---
On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance, and nothing of its path, query, headers, body or token.

## Context

A refinement of REQ-m0-exit-16-egress, which says destinations are verified from the proxy's logs; GR-12 keeps request contents and tokens out of logs. It carries this detail of REQ-e2e-egress-configured-destinations, which the firm requirement replaced. Modeled by Claude Code for the D-200 follow-ups; the author has not reviewed the wording.

## Source

> **Egress.** Outbound traffic goes only to configured destinations (IBKR's Flex service, NBP, GUS, the market-data vendor, the user's Ghostfolio), verified from the egress proxy's logs.

design-decisions:D-200#16 (docs/DESIGN.md, D-200, criterion 16); GR-12

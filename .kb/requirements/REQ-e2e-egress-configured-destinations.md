---
title: On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance and nothing of its path, query, headers, body or token, and every request it allowed must be to a host on that instance's egress allow-list.
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance and nothing of its path, query, headers, body or token, and every request it allowed must be to a host on that instance's egress allow-list.
logic_claims:
  - CLAIM-C5C852B6F860B961
semantic_clauses:
  - On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance and nothing of its path, query, headers, body or token, and every request it allowed must be to a host on that instance's egress allow-list.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e6795723bbba04918ecf318346919ce3edc2c4cc9f8574401e82561cf52b5abe
semantic_inventory:
  - claim_key: CLAIM-C5C852B6F860B961
    claim_text: On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance and nothing of its path, query, headers, body or token, and every request it allowed must be to a host on that instance's egress allow-list
    role: normative
    status: modeled
    span:
      start: 0
      end: 266
origin:
  kind: agent
  recorded_at: '2026-11-06T20:50:02.792Z'
id: REQ-e2e-egress-configured-destinations
type: req
---
On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance and nothing of its path, query, headers, body or token, and every request it allowed must be to a host on that instance's egress allow-list.

## Context

D-200 makes criterion 16 of M0's exit an end-to-end scenario verified from the egress proxy's logs, on both kind and k3d. The orchestrator's reading, given with the work package: the logs show each CONNECT with host, decision and instance and no path, query, header or body (GR-12). The suite calls no real institution (docs/dev/testing.md), so the allowed side is checked over the run's own log lines against the allow-lists.

## Source

> **Egress.** Outbound traffic goes only to configured destinations (IBKR's Flex service, NBP, GUS, the market-data vendor, the user's Ghostfolio), verified from the egress proxy's logs.

design-decisions:D-200#16 (docs/DESIGN.md, D-200, criterion 16)

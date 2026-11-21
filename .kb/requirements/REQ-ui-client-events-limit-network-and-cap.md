---
title: The limit on the pages' client event reports counts an IPv6 address as its /64 network and an IPv4 address, IPv4-mapped included, as itself, and takes at most 600 reports a minute from every address together; a report its own address's limit refuses takes nothing of that cap.
status: open
firmness: default
semantic_text: The limit on the pages' client event reports counts an IPv6 address as its /64 network and an IPv4 address, IPv4-mapped included, as itself, and takes at most 600 reports a minute from every address together; a report its own address's limit refuses takes nothing of that cap.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6db8e6632459ce604bc915cebd65332c75fc11e6e9bd62ac4b0ceec0aa41f4db
logic_claims:
  - CLAIM-C313F0EDE5F23FC8
  - CLAIM-47BF14AF15C05887
semantic_inventory:
  - claim_key: CLAIM-C313F0EDE5F23FC8
    claim_text: The limit on the pages' client event reports counts an IPv6 address as its /64 network and an IPv4 address, IPv4-mapped included, as itself, and takes at most 600 reports a minute from every address together
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 207
    subjects:
      - ui.client_events_limit_key
      - ui.client_events_shared_cap
    reason: It states what a report is counted under and a cap on all reports; no predicate of the ontology states a rate limit's key or a shared cap yet.
  - claim_key: CLAIM-47BF14AF15C05887
    claim_text: a report its own address's limit refuses takes nothing of that cap
    role: descriptive
    status: ontology_gap
    span:
      start: 209
      end: 275
    subjects:
      - ui.client_events_shared_cap
    reason: It states which reports count toward the shared cap; no predicate of the ontology states it yet.
rationale: A host with one IPv6 /64 could take a new address for every report, so the per-address limit of D-227 held it back nowhere, and it could fill the table of 4096 remembered addresses so that every other browser's reports were refused; many addresses together could flood the log and the metrics the demo exports to Grafana. The Ask box's limits already count IPv6 by /64 (D-208).
origin:
  kind: agent
  ref: Security review of the new surfaces before the alpha.2 cut, /client-events (the design session's exit criterion (c), 2026-11-21)
  recorded_at: '2026-11-21T18:15:59.408Z'
id: REQ-ui-client-events-limit-network-and-cap
type: req
---
The limit on the pages' client event reports counts an IPv6 address as its /64 network and an IPv4 address, IPv4-mapped included, as itself, and takes at most 600 reports a minute from every address together; a report its own address's limit refuses takes nothing of that cap.

## Context

A host with one IPv6 /64 could take a new address for every report, so the per-address limit of D-227 held it back nowhere, and it could fill the table of 4096 remembered addresses so that every other browser's reports were refused; many addresses together could flood the log and the metrics the demo exports to Grafana. The Ask box's limits already count IPv6 by /64 (D-208).

## Source

> Security review of the new surfaces before the alpha.2 cut: MCP, /client-events, the Turnstile cookie, the Access page and invites (the design session's exit criterion (c), 2026-11-21).

---
title: A sync's delivery must be written in one transaction whose every statement is bounded by a configurable timeout, five minutes by default, and a delivery that runs past it must fail whole, write nothing and be recorded as a failed sync attempt of the class delivery-timeout, never split.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:08.210Z'
semantic_text: A sync's delivery must be written in one transaction whose every statement is bounded by a configurable timeout, five minutes by default, and a delivery that runs past it must fail whole, write nothing and be recorded as a failed sync attempt of the class delivery-timeout, never split.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e6adb5ba37f45dd2574b5b308131655f661dd617a1b830fbb67db407074822a9
semantic_inventory:
  - claim_key: CLAIM-A2413563801B5B42
    claim_text: A sync's delivery must be written in one transaction whose every statement is bounded by a configurable timeout, five minutes by default
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 136
  - claim_key: CLAIM-F6973A47969BE390
    claim_text: a delivery that runs past it must fail whole, write nothing and be recorded as a failed sync attempt of the class delivery-timeout, never split
    role: normative
    status: ontology_gap
    span:
      start: 142
      end: 285
logic_claims:
  - CLAIM-A2413563801B5B42
  - CLAIM-F6973A47969BE390
id: REQ-ingest-delivery-timeout
type: req
---
A sync's delivery must be written in one transaction whose every statement is bounded by a configurable timeout, five minutes by default, and a delivery that runs past it must fail whole, write nothing and be recorded as a failed sync attempt of the class delivery-timeout, never split.

## Context

D-18 makes a delivery all or nothing and D-198 keeps every attempt with its error class. The timeout is a SET LOCAL statement_timeout of the delivery's transaction (--sync-delivery-timeout), so it ends with the transaction and bounds each statement, not the transaction as a whole. The class delivery-timeout says the core's database was too slow, not the connector.

## Source

> DELIVERY TIMEOUT: the sync delivery transaction (all-or-nothing, confirmed per D-18) gets a statement timeout (SET LOCAL statement_timeout inside it; value configurable, sane default e.g. 5 min). If a delivery hits it, it fails and is reported as a failed sync attempt (D-198) with a clear reason, never split silently.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
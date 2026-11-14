---
title: Every connection of the core's database pool must start with a statement timeout and an idle-in-transaction timeout, two minutes each by default and configurable through the environment, and a timeout the connection string names must win over both.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:10.440Z'
semantic_text: Every connection of the core's database pool must start with a statement timeout and an idle-in-transaction timeout, two minutes each by default and configurable through the environment, and a timeout the connection string names must win over both.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 98dc270700d0df2bd6b51d31bb653f6d685d6c0cb0146d07457b13c7ef67a91e
semantic_inventory:
  - claim_key: CLAIM-DD741692E0F86106
    claim_text: Every connection of the core's database pool must start with a statement timeout and an idle-in-transaction timeout, two minutes each by default and configurable through the environment, and a timeout the connection string names must win over both
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 247
logic_claims:
  - CLAIM-DD741692E0F86106
id: REQ-core-db-session-timeouts
type: req
---
Every connection of the core's database pool must start with a statement timeout and an idle-in-transaction timeout, two minutes each by default and configurable through the environment, and a timeout the connection string names must win over both.

## Context

The timeouts are runtime parameters of each connection, so they hold for the bundled database, the demo's and the owner's own, with no role or server setting. A connector's delivery sets its own statement timeout for its transaction. The demo's loader connects with its own connection string and sets neither.

## Source

> DB ROLE TIMEOUTS: set idle_in_transaction_session_timeout and a sane statement_timeout for the core's DB connections, configurable with defaults, in the place that also applies to the demo and the bundled DB.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
---
title: A core that issues demo sessions must, once its start's reconcile has succeeded, read its accounts and their sleeves every 15 seconds and reconcile SpiceDB when they are not empty and differ from those it last reconciled, and reconcile every five minutes whatever it read, so the accounts the demo's loader writes next to it reach SpiceDB.
status: open
firmness: default
required_evidence: unverified
semantic_text: A core that issues demo sessions must, once its start's reconcile has succeeded, read its accounts and their sleeves every 15 seconds and reconcile SpiceDB when they are not empty and differ from those it last reconciled, and reconcile every five minutes whatever it read, so the accounts the demo's loader writes next to it reach SpiceDB.
rationale: 'Inferred from cmd/demo-seed and the design session''s task: the demo switches to SpiceDB, the default (D-164), and must not break, while its loader drops and loads the ledger and the portfolio again next to a running core, which learns of no account the loader writes; a start''s reconcile that ran between the two would leave every demo session viewing nothing until the next restart.'
origin:
  kind: agent
  ref: 'the design session''s task for the SpiceDB default, 2026-11-22: Demo. It switches too. [...] The demo deploys from main automatically, so don''t let it break.'
  recorded_at: '2026-11-22T08:57:42.583Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 76c09ec6ffed3beec566728a9e12802f7e51140092858078f1e33cb1d7aa2641
semantic_inventory:
  - claim_key: CLAIM-A6547E6403ECDE51
    claim_text: A core that issues demo sessions must, once its start's reconcile has succeeded, read its accounts and their sleeves every 15 seconds and reconcile SpiceDB when they are not empty and differ from those it last reconciled, and reconcile every five minutes whatever it read, so the accounts the demo's loader writes next to it reach SpiceDB
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 338
    subjects:
      - demo.spicedb_periodic_reconcile
logic_claims:
  - CLAIM-A6547E6403ECDE51
id: REQ-demo-spicedb-reconciled-after-reseed
type: req
---
A core that issues demo sessions must, once its start's reconcile has succeeded, read its accounts and their sleeves every 15 seconds and reconcile SpiceDB when they are not empty and differ from those it last reconciled, and reconcile every five minutes whatever it read, so the accounts the demo's loader writes next to it reach SpiceDB.

## Context

The design session relayed on 2026-11-22 the author's decision that SpiceDB becomes the default authorizer before 2026.11.0-alpha.2 (D-164), and asked that the public demo switch too and not break, since it deploys from main automatically. Inferred from cmd/demo-seed (resetSQL, prepare): every deploy that changes the loader's image reseeds the demo, dropping the ledger, valuation and portfolio schemas and loading them again, while the core starts next to it (deploy/demo/openfief). Under the policy that is harmless, since GrantPolicy reads the database at every check; under SpiceDB the core writes an account's household relationship only at its start's reconcile and through its own placements, and the loader goes through neither, so a reconcile that ran after the loader's migrations and before its accounts were written removes their relationships, and demo sessions (household:home#reader@demo_session:*, D-163) view nothing. The core learns of the loader's writes from nothing else: the loader runs its own revaluation workers, writes the demo marker before it loads, and sends no notification, so the core reads the accounts and their sleeves itself, cheaply, every 15 seconds, and reconciles when they changed; an empty ledger is skipped, being only the moment between the loader's drop and its accounts, which the loader writes in its first step. That bounds the window after a reseed to about one poll plus a reconcile; the five-minute reconcile stays as a backstop. The demo's account IDs are fixed, so any later reconcile repairs it. An install without demo sessions runs neither; its accounts reach SpiceDB through the core. Reason not stated by the author beyond the demo not breaking.

## Source

> Demo. It switches too. In deploy/demo, change the comment and args so the public demo runs what owners run. [...] The demo deploys from main automatically, so don't let it break.

Source: the design session's task for the SpiceDB default, 2026-11-22.

---
title: The core must re-key each provisional instrument a sync brings the registry ISIN of, after importing the rest of the delivery, in every context that stores it with the ledger last, idempotently, refusing an invalid registry ISIN and dropping one from a connector that may not deliver bond terms.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: The core must re-key each provisional instrument a sync brings the registry ISIN of, after importing the rest of the delivery, in every context that stores it with the ledger last, idempotently, refusing an invalid registry ISIN and dropping one from a connector that may not deliver bond terms.
logic_claims:
  - CLAIM-820480C3199862C6
semantic_clauses:
  - The core must re-key each provisional instrument a sync brings the registry ISIN of, after importing the rest of the delivery, in every context that stores it with the ledger last, idempotently, refusing an invalid registry ISIN and dropping one from a connector that may not deliver bond terms.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2bea53c5afcc5df0399ded1225e2923d081e7f28f46fc93d41790a322b8a8d96
semantic_inventory:
  - claim_key: CLAIM-820480C3199862C6
    claim_text: The core must re-key each provisional instrument a sync brings the registry ISIN of, after importing the rest of the delivery, in every context that stores it with the ledger last, idempotently, refusing an invalid registry ISIN and dropping one from a connector that may not deliver bond terms
    role: normative
    status: modeled
    span:
      start: 0
      end: 294
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:17.174Z'
id: REQ-ingest-rekey-provisional-instrument
type: req
---
The core must re-key each provisional instrument a sync brings the registry ISIN of, after importing the rest of the delivery, in every context that stores it with the ledger last, idempotently, refusing an invalid registry ISIN and dropping one from a connector that may not deliver bond terms.

## Context

D-214 has a treasury bond identified by the registry's ISIN from the Ministry of Finance's data, and a lot of a series the data doesn't list yet wait for its ISIN as it waits for its terms; installs from before it hold every series under the ISIN D-99 derived from the series code, and D-161 asks that stored data be migrated. D-161 asks for a migration of stored data, and no SQL migration can know the registry's ISINs, so the next sync delivers them and the core re-keys. Modeled by Claude Code for D-214; the author has not reviewed the wording.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97), never one derived from its series code; the series code is its ticker.

docs/DESIGN.md#D-214

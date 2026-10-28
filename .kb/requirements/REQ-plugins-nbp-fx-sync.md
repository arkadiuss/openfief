---
title: The NBP FX plugin must answer Sync with NBP table A rates to PLN from the day after its watermark, or its configured start on a first sync, up to today, move the watermark to the latest delivered rate's date or keep it when none is new, reach only api.nbp.pl through the egress proxy, and refuse ImportFile.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: The NBP FX plugin must answer Sync with NBP table A rates to PLN from the day after its watermark, or its configured start on a first sync, up to today, move the watermark to the latest delivered rate's date or keep it when none is new, reach only api.nbp.pl through the egress proxy, and refuse ImportFile.
logic_claims:
  - CLAIM-05D5251DDB9E100A
semantic_clauses:
  - The NBP FX plugin must answer Sync with NBP table A rates to PLN from the day after its watermark, or its configured start on a first sync, up to today, move the watermark to the latest delivered rate's date or keep it when none is new, reach only api.nbp.pl through the egress proxy, and refuse ImportFile.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7e52709c8c52cf2776d0fcf74aa9eb713d9a667eb265192d80a8c2699791f649
semantic_inventory:
  - claim_key: CLAIM-05D5251DDB9E100A
    claim_text: The NBP FX plugin must answer Sync with NBP table A rates to PLN from the day after its watermark, or its configured start on a first sync, up to today, move the watermark to the latest delivered rate's date or keep it when none is new, reach only api.nbp.pl through the egress proxy, and refuse ImportFile
    role: normative
    status: modeled
    span:
      start: 0
      end: 306
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:23.305Z'
id: REQ-plugins-nbp-fx-sync
type: req
---
The NBP FX plugin must answer Sync with NBP table A rates to PLN from the day after its watermark, or its configured start on a first sync, up to today, move the watermark to the latest delivered rate's date or keep it when none is new, reach only api.nbp.pl through the egress proxy, and refuse ImportFile.

## Context

D-20 makes NBP table A the default FX source and a swappable FX data plugin; D-49 has the core call it on its own schedule and store the watermark it returns. A watermark at the last delivered date never skips a day NBP publishes late, and asking again is harmless because rates are recorded per currency, date and source. NBP failures are Unavailable so the core retries. Egress is only api.nbp.pl:443 (D-32). Refines REQ-AUTO-D1568676DF81CAF7 and REQ-ingest-core-calls-connectors. Modeled by Claude Code for the M0 plugin binaries task (agent A: SDK verifier, operator key publishing, first-party plugins); the author has not reviewed it yet.

## Source

> In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin.

docs/DESIGN.md D-20

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49

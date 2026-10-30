---
title: The built-in UI must show the price checks on their own page for any session, with both prices, their sources, the difference in percent, the status and each instrument's threshold, offer only the owner the acknowledgement of a flagged check and an instrument's threshold in percent, and count and link the instruments whose latest check is flagged and not acknowledged on the overview.
status: open
text_ref: docs/DESIGN.md#D-112
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-112
  - lane:strict
semantic_text: The built-in UI must show the price checks on their own page for any session, with both prices, their sources, the difference in percent, the status and each instrument's threshold, offer only the owner the acknowledgement of a flagged check and an instrument's threshold in percent, and count and link the instruments whose latest check is flagged and not acknowledged on the overview.
logic_claims:
  - CLAIM-2FF84BDA88D7699D
semantic_clauses:
  - The built-in UI must show the price checks on their own page for any session, with both prices, their sources, the difference in percent, the status and each instrument's threshold, offer only the owner the acknowledgement of a flagged check and an instrument's threshold in percent, and count and link the instruments whose latest check is flagged and not acknowledged on the overview.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 94ee61560b2912f537f3212ecb86c1d62840b299f642f961e356c80addc3386e
semantic_inventory:
  - claim_key: CLAIM-2FF84BDA88D7699D
    claim_text: The built-in UI must show the price checks on their own page for any session, with both prices, their sources, the difference in percent, the status and each instrument's threshold, offer only the owner the acknowledgement of a flagged check and an instrument's threshold in percent, and count and link the instruments whose latest check is flagged and not acknowledged on the overview
    role: normative
    status: modeled
    span:
      start: 0
      end: 385
origin:
  kind: agent
  recorded_at: '2026-10-30T05:18:14.708Z'
id: REQ-ui-price-checks-page
type: req
---
The built-in UI must show the price checks on their own page for any session, with both prices, their sources, the difference in percent, the status and each instrument's threshold, offer only the owner the acknowledgement of a flagged check and an instrument's threshold in percent, and count and link the instruments whose latest check is flagged and not acknowledged on the overview.

## Context

Restores, as a refinement, details of the superseded REQ-ui-price-checks, the overview's link among them, and adds D-110's per-instrument threshold and D-112's acknowledgement for the owner.

## Source

> The owner can acknowledge a flagged check, which is recorded append-only, and the UI shows the latest check per instrument with its history.

docs/DESIGN.md#D-112

---
title: The owner's settings page must offer the jurisdiction pack's number formats and the marginal PIT rate in percent, and the account forms each account's tax wrapper and sale fee rule with its rates in percent, handing each to the core's owner services and showing the core's reason when it refuses one.
status: open
text_ref: docs/DESIGN.md#D-78
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-78
  - lane:strict
semantic_text: The owner's settings page must offer the jurisdiction pack's number formats and the marginal PIT rate in percent, and the account forms each account's tax wrapper and sale fee rule with its rates in percent, handing each to the core's owner services and showing the core's reason when it refuses one.
logic_claims:
  - CLAIM-44405164232207EE
semantic_clauses:
  - The owner's settings page must offer the jurisdiction pack's number formats and the marginal PIT rate in percent, and the account forms each account's tax wrapper and sale fee rule with its rates in percent, handing each to the core's owner services and showing the core's reason when it refuses one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 50ed90a67e949d591f06e5841e2f412584a8108aa3016f0b73bb549091596f5b
semantic_inventory:
  - claim_key: CLAIM-44405164232207EE
    claim_text: The owner's settings page must offer the jurisdiction pack's number formats and the marginal PIT rate in percent, and the account forms each account's tax wrapper and sale fee rule with its rates in percent, handing each to the core's owner services and showing the core's reason when it refuses one
    role: normative
    status: modeled
    span:
      start: 0
      end: 299
origin:
  kind: agent
  recorded_at: '2026-10-30T02:37:47.998Z'
id: REQ-ui-value-settings-forms
type: req
---
The owner's settings page must offer the jurisdiction pack's number formats and the marginal PIT rate in percent, and the account forms each account's tax wrapper and sale fee rule with its rates in percent, handing each to the core's owner services and showing the core's reason when it refuses one.

## Context

D-78 has the owner set a fee rule per account and the marginal rate once; D-76 sets the wrapper when creating the account; D-113 makes the number format a setting. Typing rates in percent is this package's reading. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-78]** Liquidation value estimates a sale's commission with a fee rule the owner sets per account: a percentage with a minimum, plus a currency-conversion percentage for instruments in a foreign currency. The rule is zero until set. Withdrawing from IKZE before 65 is taxed as income at the owner's marginal PIT rate, a setting the owner enters once. Until it is set, an IKZE's liquidation value is shown as missing rather than guessed.

docs/DESIGN.md#D-78

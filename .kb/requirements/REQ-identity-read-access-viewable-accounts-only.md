---
title: Each holder of read access must read only the accounts it may view.
status: open
text_ref: design-decisions:D-137#6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-137-6
  - lane:strict
semantic_text: Each holder of read access must read only the accounts it may view.
logic_claims:
  - CLAIM-ACB0786EB87C0FC5
semantic_clauses:
  - Each holder of read access must read only the accounts it may view.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f314abcd59db5fee24bdd7e11f420dea3146df9e065a28322c938f6ead402072
semantic_inventory:
  - claim_key: CLAIM-ACB0786EB87C0FC5
    claim_text: Each holder of read access must read only the accounts it may view
    role: normative
    status: modeled
    span:
      start: 0
      end: 66
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-137#6
  recorded_at: '2026-11-02T00:54:01.679Z'
id: REQ-identity-read-access-viewable-accounts-only
type: req
---
Each holder of read access must read only the accounts it may view.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. Which accounts a caller may view is D-163's authorization interface's answer; until SpiceDB answers it (D-164), every holder views every account. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Read access is everything the read API serves, named by kind of record: sleeves, accounts, transactions, lots, holdings, cash, valuations with their three values, the allocation, performance, the model portfolio while it is on, price checks, balance assertion checks, transfers, attention items, yearly taxes, the settings with each account's tax wrapper, and the installed plugins. It never includes a secret, credential, passkey, session or token, a broker account number (D-121), or any change of state. The owner, script tokens, view plugins and demo sessions hold it, and nothing else does; each reads only the accounts it may view (D-163).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-137#6

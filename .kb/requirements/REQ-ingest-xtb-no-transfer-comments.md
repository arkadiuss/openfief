---
title: The XTB connector must never read, deliver or store the comment of a deposit or withdrawal, which names the sending bank or the XTB account.
status: open
text_ref: docs/research.md#R-24
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-research-md-r-24
  - lane:strict
semantic_text: The XTB connector must never read, deliver or store the comment of a deposit or withdrawal, which names the sending bank or the XTB account.
logic_claims:
  - CLAIM-CA1E365C55D6A57C
semantic_clauses:
  - The XTB connector must never read, deliver or store the comment of a deposit or withdrawal, which names the sending bank or the XTB account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: caffc9c9f14171669806d1ababd7d20928f50b940e552b7e6caae620353d7032
semantic_inventory:
  - claim_key: CLAIM-CA1E365C55D6A57C
    claim_text: The XTB connector must never read, deliver or store the comment of a deposit or withdrawal, which names the sending bank or the XTB account
    role: normative
    status: modeled
    span:
      start: 0
      end: 139
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:50.985Z'
id: REQ-ingest-xtb-no-transfer-comments
type: req
---
The XTB connector must never read, deliver or store the comment of a deposit or withdrawal, which names the sending bank or the XTB account.

## Context

R-24 says a withdrawal's comment names the XTB account number and a deposit's the sending bank; the author's work package asked that neither be stored anywhere (AGENTS.md: no account numbers in logs or data beyond what D-72 needs). Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> XTB's xStation "Historia operacji" export ... holds one account. Its labels are English, and it has three sheets, Closed Positions, Cash Operations and Open Positions.

docs/research.md#R-24

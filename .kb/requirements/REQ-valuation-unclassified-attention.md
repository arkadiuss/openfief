---
title: A revaluation of an account in a wrapper with a yearly tax must raise an attention item for each holding the pack can't classify and the owner hasn't, through the attention port.
status: open
text_ref: docs/DESIGN.md#D-132
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-132
  - lane:strict
semantic_text: A revaluation of an account in a wrapper with a yearly tax must raise an attention item for each holding the pack can't classify and the owner hasn't, through the attention port.
logic_claims:
  - CLAIM-800B6C99DCA1154F
semantic_clauses:
  - A revaluation of an account in a wrapper with a yearly tax must raise an attention item for each holding the pack can't classify and the owner hasn't, through the attention port.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 382817720e441842da943a960c001a897760eceb5d048e7880df39e0fb9c0127
semantic_inventory:
  - claim_key: CLAIM-800B6C99DCA1154F
    claim_text: A revaluation of an account in a wrapper with a yearly tax must raise an attention item for each holding the pack can't classify and the owner hasn't, through the attention port
    role: normative
    status: modeled
    span:
      start: 0
      end: 177
origin:
  kind: agent
  recorded_at: '2026-10-31T18:34:32.778Z'
id: REQ-valuation-unclassified-attention
type: req
---
A revaluation of an account in a wrapper with a yearly tax must raise an attention item for each holding the pack can't classify and the owner hasn't, through the attention port.

## Context

D-132 asks the owner once through an attention item (D-130). The attention store isn't built yet, so the core wires no port and nothing is raised (docs/implementation-questions.md). Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Source

> **[D-132]** When the pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, such as a fund whose policy it can't read, the instrument counts as neither, so the estimate errs high, and an attention item (D-130) asks the owner to classify it once; the owner's choice stays with the instrument.

docs/DESIGN.md#D-132

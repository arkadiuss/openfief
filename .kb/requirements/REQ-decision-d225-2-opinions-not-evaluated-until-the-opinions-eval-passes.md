---
title: Until the model has passed the opinions eval on its host (D-217), an answer given with opinions on names the model as not evaluated for opinions, whatever its Ask box verdict (D-222).
status: open
firmness: firm
required_evidence: asserted
tags:
  - opinions
origin:
  kind: import
  ref: design-decisions:D-225#2
  recorded_at: '2026-11-22T01:50:00.000Z'
text_ref: design-decisions:D-225#2
semantic_text: Until the model has passed the opinions eval on its host (D-217), an answer given with opinions on names the model as not evaluated for opinions, whatever its Ask box verdict (D-222).
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cb3e110324210b48e4379479ed48f0082734ba3be3ef49c1cf163883ca0b2620
semantic_inventory:
  - claim_key: CLAIM-422CE24B250BBC9E
    claim_text: Until the model has passed the opinions eval on its host (D-217), an answer given with opinions on names the model as not evaluated for opinions, whatever its Ask box verdict (D-222)
    role: descriptive
    status: missing
    span:
      start: 0
      end: 182
logic_claims:
  - CLAIM-422CE24B250BBC9E
id: REQ-decision-d225-2-opinions-not-evaluated-until-the-opinions-eval-passes
type: req
---
Until the model has passed the opinions eval on its host (D-217), an answer given with opinions on names the model as not evaluated for opinions, whatever its Ask box verdict (D-222).

## Context

Author decision D-225, as the design session recorded it (commit 128c4f6): opinions ship off by default in the second alpha with their eval set not yet approved, and a model's Ask box verdict says nothing of its opinions, so every opinions answer is labelled until the model has passed the opinions eval on its host. The note reads model names only today; its host is not in the note, so the label follows the model, and keying it by host is a follow-up.

## Source

> **[D-225]** ... Until the model has passed the opinions eval on its host (D-217), an answer given with opinions on names the model as not evaluated for opinions, whatever its Ask box verdict (D-222). ...

docs/DESIGN.md, D-225.
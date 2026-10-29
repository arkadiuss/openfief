---
title: Ingest Statement Service
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.statement_service
canonical_key: ingest.statement_service
claim_key: CLAIM-E93BA51E161C34FE
claim_text: The core must serve StatementService to view plugins only as a client stream, header first and at most 1 MiB per message, and answer the batch's counts with at most 100 ingest problems or each failure's own code.
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:38.791Z'
id: FACT-SUBJECT-8F3AFA5A57F4B4E0
type: fact
---

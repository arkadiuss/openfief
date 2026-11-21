---
title: Every save and edit of a draft thesis enters the audit log inside the transaction that stores the version, as thesis_saved or thesis_edited naming the thesis and its version and never its text, and a version the log can't record is not stored.
status: open
firmness: default
required_evidence: unverified
semantic_text: Every save and edit of a draft thesis enters the audit log inside the transaction that stores the version, as thesis_saved or thesis_edited naming the thesis and its version and never its text, and a version the log can't record is not stored.
tags:
  - opinions
origin:
  kind: agent
  ref: 'Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)'
  recorded_at: '2026-11-21T15:00:00.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e7c5f8ea6ff2832e5a8faa418fd53b8a25e69bb310e83f0d8d64289112f13d3e
logic_claims:
  - CLAIM-3489A6903060C77D
semantic_inventory:
  - claim_key: CLAIM-3489A6903060C77D
    claim_text: Every save and edit of a draft thesis enters the audit log inside the transaction that stores the version, as thesis_saved or thesis_edited naming the thesis and its version and never its text, and a version the log can't record is not stored
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 242
    subjects:
      - thesis.audit_entries
rationale: D-225 and D-229, decided by the author, as the design session's ruling of 2026-11-21 sets them out.
id: REQ-thesis-audited
type: req
---
Every save and edit of a draft thesis enters the audit log inside the transaction that stores the version, as thesis_saved or thesis_edited naming the thesis and its version and never its text, and a version the log can't record is not stored.

## Context

D-225 and D-229, decided by the author, as the design session's ruling of 2026-11-21 sets them out.

## Source

> Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)

---
title: While any audit entry's job has failed five attempts and still waits in the outbox, the core must keep one open attention item of kind audit-log-behind, in the group of data coming in, saying how many changes wait and that OpenFief keeps retrying, and must clear it when none is that stuck.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:11.587Z'
semantic_text: While any audit entry's job has failed five attempts and still waits in the outbox, the core must keep one open attention item of kind audit-log-behind, in the group of data coming in, saying how many changes wait and that OpenFief keeps retrying, and must clear it when none is that stuck.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2205519ba1bab0bf13b2077b06e355345fbfdd05508bd54c15e3d578b0db79c0
semantic_inventory:
  - claim_key: CLAIM-254F78C838661365
    claim_text: While any audit entry's job has failed five attempts and still waits in the outbox, the core must keep one open attention item of kind audit-log-behind, in the group of data coming in, saying how many changes wait and that OpenFief keeps retrying, and must clear it when none is that stuck
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 289
logic_claims:
  - CLAIM-254F78C838661365
id: REQ-attention-audit-log-behind
type: req
---
While any audit entry's job has failed five attempts and still waits in the outbox, the core must keep one open attention item of kind audit-log-behind, in the group of data coming in, saying how many changes wait and that OpenFief keeps retrying, and must clear it when none is that stuck.

## Context

D-211's outbox never drops an entry: River keeps retrying a job without limit. The count is of the jobs that failed five attempts and still wait, not of the whole backlog, so a burst of entries that only just queued raises nothing; the item clears when no job is that stuck, which is when the log has caught up. The detail changes with the count, so an acknowledged item opens again when it changes.

## Source

> AUDIT-LOG-BEHIND ATTENTION ITEM: after an audit-entry job has failed 5 attempts (keep River retrying indefinitely; never drop), raise attention item kind audit-log-behind in the data coming in group next to failing syncs (D-196). Cleared when the backlog is empty. Same state as a banner on Settings, Audit log.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
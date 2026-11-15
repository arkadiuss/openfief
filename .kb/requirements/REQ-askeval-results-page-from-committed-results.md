---
title: The Ask box evaluation page is generated on every docs build from the trimmed results of the bench's runs that are committed in the repository, one file per run, and the Ask box's note on whether a model was evaluated reads the same files.
status: open
tags:
  - askeval
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'AB9 brief: the Ask box evaluation article and the Ask box showing who answered, with the author''s D-222 (2026-11-15)'
  recorded_at: '2026-11-15T08:59:32.597Z'
semantic_text: The Ask box evaluation page is generated on every docs build from the trimmed results of the bench's runs that are committed in the repository, one file per run, and the Ask box's note on whether a model was evaluated reads the same files.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 83eb4e128185b7ec59ce0d25bf2894c38d9dbd3de6d288410aa400802268d135
semantic_inventory:
  - claim_key: CLAIM-7920278F65FB0B28
    claim_text: The Ask box evaluation page is generated on every docs build from the trimmed results of the bench's runs that are committed in the repository, one file per run, and the Ask box's note on whether a model was evaluated reads the same files
    role: descriptive
    status: missing
    span:
      start: 0
      end: 238
logic_claims:
  - CLAIM-7920278F65FB0B28
id: REQ-askeval-results-page-from-committed-results
type: req
---
The Ask box evaluation page is generated on every docs build from the trimmed results of the bench's runs that are committed in the repository, one file per run, and the Ask box's note on whether a model was evaluated reads the same files.

## Context

The Ask box links to a page that says how models did in the eval (D-222). A page typed by hand drifts from the bench's reports, so docgen renders it from committed per-run results, and the Ask box embeds the same files.

## Source

> D-222: a link to the docs page that reports how models did in the Ask box's eval (D-217): each model and host with its verdict, failing cases, reliability, cost and run dates, regenerated from the bench's reports. (The design session's build details: compact per-run result files, committed, with the bench version and configuration hash on each run.)

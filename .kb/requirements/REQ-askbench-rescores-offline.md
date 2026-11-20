---
title: The Ask box bench grades the stored answers of a report again by the current scoring rules with no model call and no key, from the answer, calls, cards and rounds the report holds, rebuilding the cards of a report that lacks them only by replaying its calls when the replay ends every call as the run did, carrying over a check it cannot grade and counting what it rebuilt, carried and kept, and writes a new report with the new scorer version into another directory.
status: open
firmness: default
semantic_text: The Ask box bench grades the stored answers of a report again by the current scoring rules with no model call and no key, from the answer, calls, cards and rounds the report holds, rebuilding the cards of a report that lacks them only by replaying its calls when the replay ends every call as the run did, carrying over a check it cannot grade and counting what it rebuilt, carried and kept, and writes a new report with the new scorer version into another directory.
tags:
  - askbench
origin:
  kind: agent
  ref: 'Design session decisions (author''s delegate): a scorer version and offline re-scoring of the Ask box bench, with D-217, D-219, D-220 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T15:38:46.874Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 58367697bfb164eb314d72226c6d804b0338a84f6a36cea3b982cfd40f0e1d5a
semantic_inventory:
  - claim_key: CLAIM-790ABDE35596EFE2
    claim_text: The Ask box bench grades the stored answers of a report again by the current scoring rules with no model call and no key, from the answer, calls, cards and rounds the report holds, rebuilding the cards of a report that lacks them only by replaying its calls when the replay ends every call as the run did, carrying over a check it cannot grade and counting what it rebuilt, carried and kept, and writes a new report with the new scorer version into another directory
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 466
    subjects:
      - ask_bench.offline_rescoring
logic_claims:
  - CLAIM-790ABDE35596EFE2
id: REQ-askbench-rescores-offline
type: req
---
The Ask box bench grades the stored answers of a report again by the current scoring rules with no model call and no key, from the answer, calls, cards and rounds the report holds, rebuilding the cards of a report that lacks them only by replaying its calls when the replay ends every call as the run did, carrying over a check it cannot grade and counting what it rebuilt, carried and kept, and writes a new report with the new scorer version into another directory.

## Context

A change of scoring rule can move the verdicts of runs that cost money to make; asking the models again would spend that money and change the answers. report.json of schema 1 held no cards' text and no output limit, so the bench now stores both and a re-grading of an old report rebuilds what it can prove and says what it could not. Decided in the design session of 2026-11-15 (D-217, D-219, D-220, D-221).

## Source

> The design session's brief to the bench work, 2026-11-15: scorer version, re-scoring with no model call, and a re-score of runs 14-1 and 14-2.
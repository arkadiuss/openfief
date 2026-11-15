---
title: The Ask box bench writes a scorer version, which is bumped whenever any scoring rule changes, into every report, trimmed result file and history entry it makes, shows it beside every verdict in its summary and on the results page, keeps it out of the configuration hash, which covers only what the model sees, and reads a report or result file without one as version 1.
status: open
firmness: default
semantic_text: The Ask box bench writes a scorer version, which is bumped whenever any scoring rule changes, into every report, trimmed result file and history entry it makes, shows it beside every verdict in its summary and on the results page, keeps it out of the configuration hash, which covers only what the model sees, and reads a report or result file without one as version 1.
tags:
  - askbench
origin:
  kind: agent
  ref: 'Design session decisions (author''s delegate): a scorer version and offline re-scoring of the Ask box bench, with D-217, D-219, D-220 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T15:38:44.378Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 58c2ac34eaa0fe6a8f3d0a059ad5be495850785fc5e2b968ce4d96bd517ead92
semantic_inventory:
  - claim_key: CLAIM-47415E2750C44B3F
    claim_text: The Ask box bench writes a scorer version, which is bumped whenever any scoring rule changes, into every report, trimmed result file and history entry it makes, shows it beside every verdict in its summary and on the results page, keeps it out of the configuration hash, which covers only what the model sees, and reads a report or result file without one as version 1
    role: descriptive
    status: missing
    span:
      start: 0
      end: 368
logic_claims:
  - CLAIM-47415E2750C44B3F
id: REQ-askbench-scorer-version
type: req
---
The Ask box bench writes a scorer version, which is bumped whenever any scoring rule changes, into every report, trimmed result file and history entry it makes, shows it beside every verdict in its summary and on the results page, keeps it out of the configuration hash, which covers only what the model sees, and reads a report or result file without one as version 1.

## Context

Runs graded by different rules are not comparable: A19 now accepts the card's change as well as the start and end values, and the same answers pass that failed before. The configuration hash must keep covering what the model sees (the model, its host and routing, its reasoning control, the prompt, the tools' schemas and the limits) so that a change of grading does not start a model's D-221 streak again. Decided in the design session of 2026-11-15 (D-217, D-219, D-220, D-221).

## Source

> The design session's brief to the bench work, 2026-11-15: scorer version, re-scoring with no model call, and a re-score of runs 14-1 and 14-2.
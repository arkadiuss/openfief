---
title: The Ask box's answer to why a connector's sync fails must say since when and for how many days the sync has been failing, how many attempts in a row failed and the latest error class, the last good sync and the next try, and list the latest attempts each with its time and its error class named as one, never its text.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box's answer to why a connector's sync fails must say since when and for how many days the sync has been failing, how many attempts in a row failed and the latest error class, the last good sync and the next try, and list the latest attempts each with its time and its error class named as one, never its text.
tags:
  - ask
origin:
  kind: agent
  ref: 'AB5 brief from the design session: bench fixes before the next run and the sync card''s error class (D-194, D-198, D-203, D-217), 2026-11-14'
  recorded_at: '2026-11-14T19:19:56.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 66b8a063a5f0172cf57121eb76cc1e7bc12969931d6cb44786ba8fc14153a1cd
semantic_inventory:
  - claim_key: CLAIM-061FA8954FCC0BD9
    claim_text: The Ask box's answer to why a connector's sync fails must say since when and for how many days the sync has been failing, how many attempts in a row failed and the latest error class, the last good sync and the next try, and list the latest attempts each with its time and its error class named as one, never its text
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 317
    subjects:
      - ask_box.sync_failure_answer
logic_claims:
  - CLAIM-061FA8954FCC0BD9
id: REQ-ui-ask-sync-answer-error-class
type: req
pair_reviews:
  - with: REQ-ui-ask-commands
    verdict: no_conflict
    separated_by: what the answer to why a sync fails must state, which this requirement fixes, against that the box understands the question as a command, which REQ-ui-ask-commands fixes
    claims:
      - claim_key: CLAIM-061FA8954FCC0BD9
    origin:
      kind: agent
      ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
---
The Ask box's answer to why a connector's sync fails must say since when and for how many days the sync has been failing, how many attempts in a row failed and the latest error class, the last good sync and the next try, and list the latest attempts each with its time and its error class named as one, never its text.

## Context

D-194 has the Ask box answer why a sync fails and D-198 keeps every attempt with its error class and never its text. The card showed the class as a bare figure, which a model read as missing data (gpt-oss: "error details unavailable", the class being unavailable) and so failed the bench's A14; it said since when the sync failed but not for how long, and not how many attempts in a row had failed or with which class. The design session decided the card is the gap and not the bench's expectation (2026-11-14).

## Source

> AB5 brief, decision 3: A14 IS A PRODUCT GAP: the sync_status card must show the latest attempts' error class and time, and how long it has been failing (D-194 "why a sync fails", D-198 "every attempt kept, with its error class, never its text").

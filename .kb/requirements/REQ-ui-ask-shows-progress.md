---
title: While a question is in flight the Ask box shows it under the field with a quiet indicator that holds still under reduced motion and says what it is doing (Looking it up, Thinking, Checking you're a person), disables its Ask button, and then shows the answer directly under the field and in view, with the help folded behind How this works and the suggestions in one row that scrolls on its own at phone width.
status: open
semantic_text: While a question is in flight the Ask box shows it under the field with a quiet indicator that holds still under reduced motion and says what it is doing (Looking it up, Thinking, Checking you're a person), disables its Ask button, and then shows the answer directly under the field and in view, with the help folded behind How this works and the suggestions in one row that scrolls on its own at phone width.
origin:
  kind: agent
  ref: 'Author feedback on demo.openfief.dev from Brave for Android (2026-11-19): a token reported ready but never sent, the check re-running, no progress, too much text; the design session approved the on-demand check and the single attribution note, D-208 and D-222 unchanged'
  recorded_at: '2026-11-19'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: da8e15d9ba422ae3955182120b4ac0aaad02ee9543d5914234c5e47f685d2d0c
semantic_inventory:
  - claim_key: CLAIM-CEFBFDFBA01C670A
    claim_text: While a question is in flight the Ask box shows it under the field with a quiet indicator that holds still under reduced motion and says what it is doing (Looking it up, Thinking, Checking you're a person), disables its Ask button, and then shows the answer directly under the field and in view, with the help folded behind How this works and the suggestions in one row that scrolls on its own at phone width
    role: descriptive
    status: missing
    span:
      start: 0
      end: 408
logic_claims:
  - CLAIM-CEFBFDFBA01C670A
id: REQ-ui-ask-shows-progress
type: req
---
While a question is in flight the Ask box shows it under the field with a quiet indicator that holds still under reduced motion and says what it is doing (Looking it up, Thinking, Checking you're a person), disables its Ask button, and then shows the answer directly under the field and in view, with the help folded behind How this works and the suggestions in one row that scrolls on its own at phone width.

## Context

The author, on Brave for Android: no progress while the model works, lots of text that is not needed, big stacked chips take the whole screen and the answer appears far below. The progress block is shown by htmx's own request class and the check's wait by a class of ask.js, so the strict policy needs no inline script or style.

## Source

> The author's feedback on demo.openfief.dev (2026-11-19): no progress, lots of not needed text, on mobile doesn't feel right.

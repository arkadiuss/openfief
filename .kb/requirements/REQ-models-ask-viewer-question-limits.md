---
title: A viewer's session must be limited, by default, to 10 questions an hour and 30 a day to a model per session and per address, besides any limit set for every session, the stricter of each window holding, while the owner's own questions stay unlimited unless the owner sets a limit for everyone; the owner can lift or change a viewer's limits by flag.
status: open
firmness: default
semantic_text: A viewer's session must be limited, by default, to 10 questions an hour and 30 a day to a model per session and per address, besides any limit set for every session, the stricter of each window holding, while the owner's own questions stay unlimited unless the owner sets a limit for everyone; the owner can lift or change a viewer's limits by flag.
origin:
  kind: agent
  ref: Viewers' per-session question limits are on by default. Pick a conservative default for viewer sessions only, such as 10 an hour and 30 a day, and leave the owner's unlimited unless set. (the owner's items for alpha.2, from the reviews of main 0bfeb62)
  recorded_at: '2026-11-27T08:18:11.302Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: db6eadcc4b6ebcdd56e0dbb5e48397e60436c8d99ea026e693fed0176e7cfc0b
semantic_inventory:
  - claim_key: CLAIM-D01773D8FE898537
    claim_text: A viewer's session must be limited, by default, to 10 questions an hour and 30 a day to a model per session and per address, besides any limit set for every session, the stricter of each window holding, while the owner's own questions stay unlimited unless the owner sets a limit for everyone
    role: exception
    status: missing
    span:
      start: 0
      end: 292
    subjects:
      - models.ask_viewer_question_limits
  - claim_key: CLAIM-6F156FA430F78C5C
    claim_text: the owner can lift or change a viewer's limits by flag
    role: descriptive
    status: missing
    span:
      start: 294
      end: 348
logic_claims:
  - CLAIM-D01773D8FE898537
  - CLAIM-6F156FA430F78C5C
id: REQ-models-ask-viewer-question-limits
type: req
---
A viewer's session must be limited, by default, to 10 questions an hour and 30 a day to a model per session and per address, besides any limit set for every session, the stricter of each window holding, while the owner's own questions stay unlimited unless the owner sets a limit for everyone; the owner can lift or change a viewer's limits by flag.

## Context

Every question a viewer asks is a model call the owner pays for, and the question limits were off unless set, so a viewer invited to an install with a model on could spend its budget without end (finding S3 of the reviews of main 0bfeb62). D-208's 10 an hour and 30 a day are the conservative default; REQ-models-ask-question-limits keeps the owner's unlimited unless set.

## Source

> Viewers' per-session question limits are on by default. Pick a conservative default for viewer sessions only, such as 10 an hour and 30 a day, and leave the owner's unlimited unless set. (the owner's items for alpha.2, from the reviews of main 0bfeb62)
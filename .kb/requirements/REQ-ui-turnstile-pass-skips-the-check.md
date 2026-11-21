---
title: A model question with a valid pass goes on with no token and no call to Cloudflare's siteverify, while the per-session and per-address limits and the day's budget still count it like any other model question, and a command never needs the pass.
status: open
firmness: default
required_evidence: unverified
semantic_text: A model question with a valid pass goes on with no token and no call to Cloudflare's siteverify, while the per-session and per-address limits and the day's budget still count it like any other model question, and a command never needs the pass.
origin:
  kind: agent
  ref: 'The design session''s amendment of D-208 (docs/DESIGN.md): a visitor who passes Turnstile isn''t checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question (task brief, 2026-11-21)'
  recorded_at: '2026-11-21T12:30:00.000Z'
tags:
  - turnstile
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fb35f33931933da788949ea6629dd2e1f0fc41730a9153db550cb582be9fc278
logic_claims:
  - CLAIM-575E087C127F3780
semantic_inventory:
  - claim_key: CLAIM-575E087C127F3780
    claim_text: A model question with a valid pass goes on with no token and no call to Cloudflare's siteverify, while the per-session and per-address limits and the day's budget still count it like any other model question, and a command never needs the pass
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 243
    subjects:
      - turnstile.pass_skips_check
    reason: It states how the built-in UI behaves; no predicate of the ontology states it yet.
rationale: 'D-208, as amended: the rate limits and the budget still count every question. The pass skips only the check that a person is asking; the core''s admission comes after it as before, and a command never reached the check.'
id: REQ-ui-turnstile-pass-skips-the-check
type: req
---
A model question with a valid pass goes on with no token and no call to Cloudflare's siteverify, while the per-session and per-address limits and the day's budget still count it like any other model question, and a command never needs the pass.

## Context

D-208, as amended: the rate limits and the budget still count every question. The pass skips only the check that a person is asking; the core's admission comes after it as before, and a command never reached the check.

## Source

> D-208 (docs/DESIGN.md), as amended: A visitor who passes Turnstile isn't checked again for 30 minutes after their last model question, and at most 2 hours from the check, through a signed cookie tied to their session that holds nothing else; the rate limits and the budget still count every question.

---
title: 'When the Ask box''s check stops a question, its card says why: no token reached OpenFief, Cloudflare refused the token (with Cloudflare''s error codes), or the token was made for another host (naming both hosts); when the check cannot finish in the browser, one message in the answer area says so with Cloudflare''s error code or a fixed word (blocked, timeout, unsupported) and what to do, with the hint to allow challenges.cloudflare.com (in Brave: turn Shields off), and the box has no status line; no token, secret or question is ever shown or logged.'
status: open
firmness: default
semantic_text: 'When the Ask box''s check stops a question, its card says why: no token reached OpenFief, Cloudflare refused the token (with Cloudflare''s error codes), or the token was made for another host (naming both hosts); when the check cannot finish in the browser, one message in the answer area says so with Cloudflare''s error code or a fixed word (blocked, timeout, unsupported) and what to do, with the hint to allow challenges.cloudflare.com (in Brave: turn Shields off), and the box has no status line; no token, secret or question is ever shown or logged.'
origin:
  kind: agent
  ref: 'Author feedback on demo.openfief.dev from Brave for Android (2026-11-19): Ready but no token, re-verifying, no progress, too much text; design session approved the on-demand check (D-208 unchanged)'
  recorded_at: '2026-11-19'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 84daef8db1184538df82a827a7cf8fff5dd7a7ca087353a75a9392ba9a0db787
semantic_inventory:
  - claim_key: CLAIM-BDD32E7B2C8802B9
    claim_text: 'When the Ask box''s check stops a question, its card says why: no token reached OpenFief, Cloudflare refused the token (with Cloudflare''s error codes), or the token was made for another host (naming both hosts)'
    role: condition
    status: ontology_gap
    span:
      start: 0
      end: 209
  - claim_key: CLAIM-DBA616B308E26DAD
    claim_text: 'when the check cannot finish in the browser, one message in the answer area says so with Cloudflare''s error code or a fixed word (blocked, timeout, unsupported) and what to do, with the hint to allow challenges.cloudflare.com (in Brave: turn Shields off), and the box has no status line'
    role: condition
    status: missing
    span:
      start: 211
      end: 497
  - claim_key: CLAIM-7361B046815DCA1B
    claim_text: no token, secret or question is ever shown or logged
    role: descriptive
    status: missing
    span:
      start: 499
      end: 551
logic_claims:
  - CLAIM-BDD32E7B2C8802B9
  - CLAIM-DBA616B308E26DAD
  - CLAIM-7361B046815DCA1B
id: REQ-ui-turnstile-failures-visible
type: req
---
When the Ask box's check stops a question, its card says why: no token reached OpenFief, Cloudflare refused the token (with Cloudflare's error codes), or the token was made for another host (naming both hosts); when the check cannot finish in the browser, one message in the answer area says so with Cloudflare's error code or a fixed word (blocked, timeout, unsupported) and what to do, with the hint to allow challenges.cloudflare.com (in Brave: turn Shields off), and the box has no status line; no token, secret or question is ever shown or logged.

## Context

The owner turned on the demo's model and every free question stopped with one card, while the widget, hidden by interaction-only, showed nothing. The demo's logs cannot be read, so the page must say why. The reasons shown are Cloudflare's fixed error codes and host names, never secrets (D-208).

## Source

> The author's report on demo.openfief.dev (2026-11-18): every free question answers "A check comes before the model answers" and no Turnstile widget is ever visible; make every failure visible and specific without weakening the check.

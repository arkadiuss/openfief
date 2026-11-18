---
title: 'When the Ask box''s check stops a question, its card says why: no token reached OpenFief, Cloudflare refused the token (with Cloudflare''s error codes), or the token was made for another host (naming both hosts); a status line in the form, a polite live region, says the check is running until a token arrives, is ready once it has and shows the widget''s error code when it fails, so a pending or failed interaction-only widget is never invisible; no token, secret or question is ever shown or logged.'
status: open
firmness: default
semantic_text: 'When the Ask box''s check stops a question, its card says why: no token reached OpenFief, Cloudflare refused the token (with Cloudflare''s error codes), or the token was made for another host (naming both hosts); a status line in the form, a polite live region, says the check is running until a token arrives, is ready once it has and shows the widget''s error code when it fails, so a pending or failed interaction-only widget is never invisible; no token, secret or question is ever shown or logged.'
origin:
  kind: agent
  ref: Ask box check failures made visible and specific, D-208, 2026-11-18
  recorded_at: '2026-11-18T15:15:42.031Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5d3f6a5eb02166ef6045e7b26ae0568b23f8b8bd7e5c00ce32c4be85a953a94a
semantic_inventory:
  - claim_key: CLAIM-BDD32E7B2C8802B9
    claim_text: 'When the Ask box''s check stops a question, its card says why: no token reached OpenFief, Cloudflare refused the token (with Cloudflare''s error codes), or the token was made for another host (naming both hosts)'
    role: condition
    status: ontology_gap
    span:
      start: 0
      end: 209
  - claim_key: CLAIM-BF1492941E526259
    claim_text: a status line in the form, a polite live region, says the check is running until a token arrives, is ready once it has and shows the widget's error code when it fails, so a pending or failed interaction-only widget is never invisible
    role: normative
    status: ontology_gap
    span:
      start: 211
      end: 444
  - claim_key: CLAIM-7361B046815DCA1B
    claim_text: no token, secret or question is ever shown or logged
    role: descriptive
    status: missing
    span:
      start: 446
      end: 498
logic_claims:
  - CLAIM-BDD32E7B2C8802B9
  - CLAIM-BF1492941E526259
  - CLAIM-7361B046815DCA1B
id: REQ-ui-turnstile-failures-visible
type: req
---
When the Ask box's check stops a question, its card says why: no token reached OpenFief, Cloudflare refused the token (with Cloudflare's error codes), or the token was made for another host (naming both hosts); a status line in the form, a polite live region, says the check is running until a token arrives, is ready once it has and shows the widget's error code when it fails, so a pending or failed interaction-only widget is never invisible; no token, secret or question is ever shown or logged.

## Context

The owner turned on the demo's model and every free question stopped with one card, while the widget, hidden by interaction-only, showed nothing. The demo's logs cannot be read, so the page must say why. The reasons shown are Cloudflare's fixed error codes and host names, never secrets (D-208).

## Source

> The author's report on demo.openfief.dev (2026-11-18): every free question answers "A check comes before the model answers" and no Turnstile widget is ever visible; make every failure visible and specific without weakening the check.

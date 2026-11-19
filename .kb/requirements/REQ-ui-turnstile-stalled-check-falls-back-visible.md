---
title: When the Ask box's invisible Turnstile check, run because a model question needs it, gives no token, no click request and no error for about eight seconds, the page renders the widget again as a visible checkbox and says to complete the check below, the checks that follow on that page start with the visible widget, and a check with no token 15 seconds after that is called timed out in one message with the code timeout; which path completed, invisible or visible fallback, is logged beside the verdict as one of two fixed words and nothing typed is ever logged or shown.
status: open
firmness: default
origin:
  kind: agent
  ref: 'Author feedback on demo.openfief.dev from Brave for Android (2026-11-19): a token reported ready but never sent, the check re-running, no progress, too much text; the design session approved the on-demand check and the single attribution note, D-208 and D-222 unchanged'
  recorded_at: '2026-11-19'
id: REQ-ui-turnstile-stalled-check-falls-back-visible
type: req
semantic_text: When the Ask box's invisible Turnstile check, run because a model question needs it, gives no token, no click request and no error for about eight seconds, the page renders the widget again as a visible checkbox and says to complete the check below, the checks that follow on that page start with the visible widget, and a check with no token 15 seconds after that is called timed out in one message with the code timeout; which path completed, invisible or visible fallback, is logged beside the verdict as one of two fixed words and nothing typed is ever logged or shown.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2ee590316fbb81df901c8398841a2dddd3b9adc1a4e57074c9ff1b4c4f5bb05e
semantic_inventory:
  - claim_key: CLAIM-F99B99F99FA067D0
    claim_text: When the Ask box's invisible Turnstile check, run because a model question needs it, gives no token, no click request and no error for about eight seconds, the page renders the widget again as a visible checkbox and says to complete the check below, the checks that follow on that page start with the visible widget, and a check with no token 15 seconds after that is called timed out in one message with the code timeout
    role: rationale
    status: nonlogical
    span:
      start: 0
      end: 421
  - claim_key: CLAIM-AAE9714B4F103E02
    claim_text: which path completed, invisible or visible fallback, is logged beside the verdict as one of two fixed words and nothing typed is ever logged or shown
    role: descriptive
    status: missing
    span:
      start: 423
      end: 572
logic_claims:
  - CLAIM-AAE9714B4F103E02
---
When the Ask box's invisible Turnstile check, run because a model question needs it, gives no token, no click request and no error for about eight seconds, the page renders the widget again as a visible checkbox and says to complete the check below, the checks that follow on that page start with the visible widget, and a check with no token 15 seconds after that is called timed out in one message with the code timeout; which path completed, invisible or visible fallback, is logged beside the verdict as one of two fixed words and nothing typed is ever logged or shown.

## Context

On Brave for Android the interaction-only challenge never completes: the script loads but no token, no before-interactive callback and no error arrive, so the status line says "Checking you're a person…" for ever. The design session approved a pure UX fix with no decision change to D-208: a stalled invisible check becomes the visible managed widget a person can tick. The path that completed is logged so the owner learns how often the fallback is the one that works.

## Source

> The design session's approval (2026-11-18): if no token has arrived about 10 s after the widget rendered, re-render it as the visible widget with the same callbacks; log which path completed with fixed words only, validated against that set and never echoed.

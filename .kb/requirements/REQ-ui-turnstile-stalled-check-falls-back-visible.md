---
title: When the Ask box's invisible Turnstile check gives no token, no click request and no error for about ten seconds, the page renders the widget again as a visible checkbox and says to complete the check below; which path completed, invisible or visible fallback, is logged beside the verdict as one of two fixed words and nothing typed is ever logged or shown.
status: open
firmness: default
origin:
  kind: agent
  ref: 'Design session: Turnstile never completes on Brave for Android, D-208 (pure UX), 2026-11-18'
  recorded_at: '2026-11-18T21:25:31.967Z'
id: REQ-ui-turnstile-stalled-check-falls-back-visible
type: req
semantic_text: When the Ask box's invisible Turnstile check gives no token, no click request and no error for about ten seconds, the page renders the widget again as a visible checkbox and says to complete the check below; which path completed, invisible or visible fallback, is logged beside the verdict as one of two fixed words and nothing typed is ever logged or shown.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d2d36c46370d495b9a2e1fcaf08ba270cc0329672c4c4ec44d2ba4c7adc177d8
semantic_inventory:
  - claim_key: CLAIM-353EC195CB14E705
    claim_text: When the Ask box's invisible Turnstile check gives no token, no click request and no error for about ten seconds, the page renders the widget again as a visible checkbox and says to complete the check below
    role: condition
    status: ontology_gap
    span:
      start: 0
      end: 206
  - claim_key: CLAIM-AAE9714B4F103E02
    claim_text: which path completed, invisible or visible fallback, is logged beside the verdict as one of two fixed words and nothing typed is ever logged or shown
    role: descriptive
    status: missing
    span:
      start: 208
      end: 357
logic_claims:
  - CLAIM-353EC195CB14E705
  - CLAIM-AAE9714B4F103E02
---
When the Ask box's invisible Turnstile check gives no token, no click request and no error for about ten seconds, the page renders the widget again as a visible checkbox and says to complete the check below; which path completed, invisible or visible fallback, is logged beside the verdict as one of two fixed words and nothing typed is ever logged or shown.

## Context

On Brave for Android the interaction-only challenge never completes: the script loads but no token, no before-interactive callback and no error arrive, so the status line says "Checking you're a person…" for ever. The design session approved a pure UX fix with no decision change to D-208: a stalled invisible check becomes the visible managed widget a person can tick. The path that completed is logged so the owner learns how often the fallback is the one that works.

## Source

> The design session's approval (2026-11-18): if no token has arrived about 10 s after the widget rendered, re-render it as the visible widget with the same callbacks; log which path completed with fixed words only, validated against that set and never echoed.
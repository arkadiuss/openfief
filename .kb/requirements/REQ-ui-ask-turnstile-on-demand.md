---
title: 'The Ask box checks a person only when a model question needs it: every question goes first without a token and a command is answered at once; the response to a model question that came with no token, and no other response, carries the HX-Trigger event openfief-check-needed; the page then runs the widget with execute, shows progress, asks the same question again with the token it made, keeps no token between questions and never checks again by itself; and the server still stops every model question that has neither a token Cloudflare accepts nor a valid pass (D-208).'
status: open
semantic_text: 'The Ask box checks a person only when a model question needs it: every question goes first without a token and a command is answered at once; the response to a model question that came with no token, and no other response, carries the HX-Trigger event openfief-check-needed; the page then runs the widget with execute, shows progress, asks the same question again with the token it made, keeps no token between questions and never checks again by itself; and the server still stops every model question that has neither a token Cloudflare accepts nor a valid pass (D-208).'
origin:
  kind: agent
  ref: 'Author feedback on demo.openfief.dev from Brave for Android (2026-11-19): a token reported ready but never sent, the check re-running, no progress, too much text; the design session approved the on-demand check and the single attribution note, D-208 and D-222 unchanged'
  recorded_at: '2026-11-19'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 74430f6c3c2da0f7e10297ede7d80485bfc73bc93c60ad17b0372145614e396f
semantic_inventory:
  - claim_key: CLAIM-63978BC89C7652C0
    claim_text: 'The Ask box checks a person only when a model question needs it: every question goes first without a token and a command is answered at once'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 140
    subjects:
      - ask_box.turnstile_on_demand
  - claim_key: CLAIM-CC9E5F073242BB82
    claim_text: the response to a model question that came with no token, and no other response, carries the HX-Trigger event openfief-check-needed
    role: descriptive
    status: missing
    span:
      start: 142
      end: 273
  - claim_key: CLAIM-2F2E7436EB78A497
    claim_text: the page then runs the widget with execute, shows progress, asks the same question again with the token it made, keeps no token between questions and never checks again by itself
    role: descriptive
    status: missing
    span:
      start: 275
      end: 453
  - claim_key: CLAIM-8852879E60EAA192
    claim_text: and the server still stops every model question that has neither a token Cloudflare accepts nor a valid pass (D-208)
    role: descriptive
    status: missing
    span:
      start: 455
      end: 571
logic_claims:
  - CLAIM-63978BC89C7652C0
  - CLAIM-CC9E5F073242BB82
  - CLAIM-2F2E7436EB78A497
  - CLAIM-8852879E60EAA192
id: REQ-ui-ask-turnstile-on-demand
type: req
---
The Ask box checks a person only when a model question needs it: every question goes first without a token and a command is answered at once; the response to a model question that came with no token, and no other response, carries the HX-Trigger event openfief-check-needed; the page then runs the widget with execute, shows progress, asks the same question again with the token it made, keeps no token between questions and never checks again by itself; and the server still stops every model question that has neither a token Cloudflare accepts nor a valid pass (D-208).

## Context

The Ask box used to make a token in advance and renew it after every question. On Brave for Android the author saw the line say Ready while the question came back with the no-token card, and the visible fallback widget re-ran during normal use. The cause was the pre-fetched token: each question spends the one it holds, and its replacement takes seconds (longer where Cloudflare retries), so a question asked in that window, or a second question straight after the first, left with no token, while the line, which describes the token now and not the token at the moment of asking, was Ready again by the time the card was read. The widget that was reset after every request also re-ran the challenge, which on a stalled browser meant the person watching Cloudflare check them again and again. Running the check when a model question needs it removes both: no token waits, nothing re-runs, and the question that goes again is the one that was asked. D-208 is unchanged by the on-demand check: the server re-validates every model question before anything is sent or counted, on a token Cloudflare accepts or on a valid pass, the signed cookie a visitor who passed the check holds for 30 minutes after their last model question and at most 2 hours from the check (D-208 as amended; REQ-ui-turnstile-pass-window, REQ-ui-turnstile-pass-cookie and REQ-ui-turnstile-pass-skips-the-check own the pass). Amended 2026-11-21 in place: this requirement is of default firmness, and its statement used to say that every model question needs a token Cloudflare accepts, which the pass made untrue.

## Source

> The design session's approval (2026-11-19): turnstile on demand is approved, D-208 unchanged; the author's report on demo.openfief.dev from Brave for Android: Ready but no token, retried multiple times, no progress, too much text.

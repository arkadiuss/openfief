---
title: When the installation configures Cloudflare Turnstile, the Ask box shows its widget beside the field only where a model answers, loads the widget's script and frame from Cloudflare's origin and nothing else from outside, with the UI's own content security policy naming that one origin and nothing loose, makes a new token after each question, needs none for a command, and, without the configuration, shows no widget and loads nothing from Cloudflare.
status: open
firmness: default
semantic_text: When the installation configures Cloudflare Turnstile, the Ask box shows its widget beside the field only where a model answers, loads the widget's script and frame from Cloudflare's origin and nothing else from outside, with the UI's own content security policy naming that one origin and nothing loose, makes a new token after each question, needs none for a command, and, without the configuration, shows no widget and loads nothing from Cloudflare.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T22:46:58.552Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 10aa4e39eed907700185a762a6e3d7a77bb4fb9f9056aafdbf6eb25926088eb5
semantic_inventory:
  - claim_key: CLAIM-7F749D03E235F6EA
    claim_text: When the installation configures Cloudflare Turnstile, the Ask box shows its widget beside the field only where a model answers, loads the widget's script and frame from Cloudflare's origin and nothing else from outside, with the UI's own content security policy naming that one origin and nothing loose, makes a new token after each question, needs none for a command, and, without the configuration, shows no widget and loads nothing from Cloudflare
    role: condition
    status: ontology_gap
    span:
      start: 0
      end: 451
logic_claims:
  - CLAIM-7F749D03E235F6EA
id: REQ-ui-ask-turnstile
type: req
---
When the installation configures Cloudflare Turnstile, the Ask box shows its widget beside the field only where a model answers, loads the widget's script and frame from Cloudflare's origin and nothing else from outside, with the UI's own content security policy naming that one origin and nothing loose, makes a new token after each question, needs none for a command, and, without the configuration, shows no widget and loads nothing from Cloudflare.

## Context

D-208's fourth guard: Cloudflare Turnstile in front of the demo's Ask box, the only third-party script the demo allows. It is configured, so self-hosters are unaffected, and its origin is the one outside host the page's policy may name. A token is good for one check, so the widget renews it after every request.

## Source

> The design session's M5 corrections (2026-11-14): Cloudflare Turnstile in front of the demo's Ask box: the widget on the Ask page only when configured (site key from config), CSP updated to allow only Turnstile's script/frame on the demo (the only third-party script allowed), ui:browser with Turnstile off and a fake on (M5 brief, 2026-11-14).
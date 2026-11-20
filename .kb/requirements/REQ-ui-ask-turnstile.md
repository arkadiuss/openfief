---
title: When the installation configures Cloudflare Turnstile, the Ask box loads the widget's script from Cloudflare's origin and nothing else from outside, with the UI's own content security policy naming that one origin and nothing loose, only where a model answers, renders the widget on demand in the interaction-only form, outside the form, never sends a token with a later question, needs no check for a command, and, without the configuration, shows no widget and loads nothing from Cloudflare.
status: open
firmness: default
semantic_text: When the installation configures Cloudflare Turnstile, the Ask box loads the widget's script from Cloudflare's origin and nothing else from outside, with the UI's own content security policy naming that one origin and nothing loose, only where a model answers, renders the widget on demand in the interaction-only form, outside the form, never sends a token with a later question, needs no check for a command, and, without the configuration, shows no widget and loads nothing from Cloudflare.
origin:
  kind: agent
  ref: 'Author feedback on demo.openfief.dev from Brave for Android (2026-11-19): a token reported ready but never sent, the check re-running, no progress, too much text; the design session approved the on-demand check and the single attribution note, D-208 and D-222 unchanged'
  recorded_at: '2026-11-19'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 22c7b0a4221e2db815843e9f67b325a871b49910b7d8a74f16a2e1a048f163f8
semantic_inventory:
  - claim_key: CLAIM-E269FAF131F75C5E
    claim_text: When the installation configures Cloudflare Turnstile, the Ask box loads the widget's script from Cloudflare's origin and nothing else from outside, with the UI's own content security policy naming that one origin and nothing loose, only where a model answers, renders the widget on demand in the interaction-only form, outside the form, never sends a token with a later question, needs no check for a command, and, without the configuration, shows no widget and loads nothing from Cloudflare
    role: condition
    status: ontology_gap
    span:
      start: 0
      end: 492
logic_claims:
  - CLAIM-E269FAF131F75C5E
id: REQ-ui-ask-turnstile
type: req
---
When the installation configures Cloudflare Turnstile, the Ask box loads the widget's script from Cloudflare's origin and nothing else from outside, with the UI's own content security policy naming that one origin and nothing loose, only where a model answers, renders the widget on demand in the interaction-only form, outside the form, never sends a token with a later question, needs no check for a command, and, without the configuration, shows no widget and loads nothing from Cloudflare.

## Context

D-208's fourth guard: Cloudflare Turnstile in front of the demo's Ask box, the only third-party script the demo allows. It is configured, so self-hosters are unaffected, and its origin is the one outside host the page's policy may name. A token is good for one check, so the widget renews it after every request.

## Source

> The design session's M5 corrections (2026-11-14): Cloudflare Turnstile in front of the demo's Ask box: the widget on the Ask page only when configured (site key from config), CSP updated to allow only Turnstile's script/frame on the demo (the only third-party script allowed), ui:browser with Turnstile off and a fake on (M5 brief, 2026-11-14).

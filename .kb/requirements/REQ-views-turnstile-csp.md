---
title: With --allow-turnstile the content security policy the core adds to every page it proxies, and to its own refusals, also allows scripts and frames from Cloudflare Turnstile's origin and from no other host beyond the core's own, and without the flag the policy is the strict one it always was.
status: open
firmness: default
semantic_text: With --allow-turnstile the content security policy the core adds to every page it proxies, and to its own refusals, also allows scripts and frames from Cloudflare Turnstile's origin and from no other host beyond the core's own, and without the flag the policy is the strict one it always was.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:41:40.826Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b7265120bc5a5cecf0cb9e9e5e94b527c4d028dea02cda3cd99299e126a09d91
semantic_inventory:
  - claim_key: CLAIM-26A30B3AD7F9ED36
    claim_text: With --allow-turnstile the content security policy the core adds to every page it proxies, and to its own refusals, also allows scripts and frames from Cloudflare Turnstile's origin and from no other host beyond the core's own, and without the flag the policy is the strict one it always was
    role: descriptive
    status: missing
    span:
      start: 0
      end: 291
logic_claims:
  - CLAIM-26A30B3AD7F9ED36
id: REQ-views-turnstile-csp
type: req
---
With --allow-turnstile the content security policy the core adds to every page it proxies, and to its own refusals, also allows scripts and frames from Cloudflare Turnstile's origin and from no other host beyond the core's own, and without the flag the policy is the strict one it always was.

## Context

A browser enforces every policy a page carries, so the built-in UI's own policy can't allow a script the core's policy blocks: the core's default-src 'self' would stop the Turnstile widget. D-208 makes Turnstile the only third-party script the demo allows, so the widened policy names its one origin, challenges.cloudflare.com, in script-src and frame-src and nothing else, and only where the installation turns it on.

## Source

> The design session's M5 corrections: CSP updated to allow only Turnstile's script/frame on the demo, the only third-party script allowed; off unless configured, self-hosters unaffected (M5 brief, 2026-11-14).
---
title: A question the Ask box sends a model passes only with a token that Cloudflare's siteverify accepts for the host that is asking, checked before the core counts the question, so a bot that passes no check spends no one's allowance; no token or a refused one stops the question with the box's own words, a check that can't be made (Cloudflare unreachable, an answer that is no JSON, a secret Cloudflare refuses or the UI can't read) stops it too and sends nothing, the secret key comes from the instance's Secret on each call and neither it, a token nor a question is logged, and the plugin's config sets the site key and siteverify URL together, whose host is the one egress host its manifest asks for.
status: open
firmness: default
semantic_text: A question the Ask box sends a model passes only with a token that Cloudflare's siteverify accepts for the host that is asking, checked before the core counts the question, so a bot that passes no check spends no one's allowance; no token or a refused one stops the question with the box's own words, a check that can't be made (Cloudflare unreachable, an answer that is no JSON, a secret Cloudflare refuses or the UI can't read) stops it too and sends nothing, the secret key comes from the instance's Secret on each call and neither it, a token nor a question is logged, and the plugin's config sets the site key and siteverify URL together, whose host is the one egress host its manifest asks for.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T22:47:01.340Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a7351566b73a3431b546799fb1a35868986ff518741f131549786ec67102120b
semantic_inventory:
  - claim_key: CLAIM-78E18D7F2F8CAF0F
    claim_text: A question the Ask box sends a model passes only with a token that Cloudflare's siteverify accepts for the host that is asking, checked before the core counts the question, so a bot that passes no check spends no one's allowance
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 228
    subjects:
      - turnstile.token_verification
  - claim_key: CLAIM-D8FEC2CA978BA638
    claim_text: no token or a refused one stops the question with the box's own words, a check that can't be made (Cloudflare unreachable, an answer that is no JSON, a secret Cloudflare refuses or the UI can't read) stops it too and sends nothing, the secret key comes from the instance's Secret on each call and neither it, a token nor a question is logged, and the plugin's config sets the site key and siteverify URL together, whose host is the one egress host its manifest asks for
    role: normative
    status: missing
    span:
      start: 230
      end: 699
    subjects:
      - turnstile.token_verification
logic_claims:
  - CLAIM-78E18D7F2F8CAF0F
  - CLAIM-D8FEC2CA978BA638
id: REQ-ui-turnstile-verification
type: req
---
A question the Ask box sends a model passes only with a token that Cloudflare's siteverify accepts for the host that is asking, checked before the core counts the question, so a bot that passes no check spends no one's allowance; no token or a refused one stops the question with the box's own words, a check that can't be made (Cloudflare unreachable, an answer that is no JSON, a secret Cloudflare refuses or the UI can't read) stops it too and sends nothing, the secret key comes from the instance's Secret on each call and neither it, a token nor a question is logged, and the plugin's config sets the site key and siteverify URL together, whose host is the one egress host its manifest asks for.

## Context

The widget runs in the visitor's browser, so only a server-side check can tell a person from a script. The core has no egress, so the UI plugin, which does, calls Cloudflare, with the one host its grant approves (D-75) and the secret in a Secret created out of band, never in git (D-29). Failing closed keeps a Cloudflare outage from becoming an open door to the key.

## Source

> The design session's M5 corrections (2026-11-14): Server-side verification of the token before a question goes to the model (siteverify call from the core or the UI plugin, whichever has egress; respect D-75 egress rules: add challenges.cloudflare.com egress only where needed); the secret key from a Secret by name, never in git; tests with a fake siteverify server (M5 brief, 2026-11-14).
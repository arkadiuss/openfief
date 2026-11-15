---
title: The demo's model layer carries D-208's spend guards, as the core's flags (a positive daily budget, 10 questions an hour and 30 a day per session and per address, 3 calls at once, the trusted proxy of the address header and Turnstile's place in the policy) and the built-in UI's config and grant (the limits of one question, a site key, the verification host as an egress host and the Secret ui--turnstile-secret), checked by validate.sh, which also fails a live demo that includes the layer while the site key is the placeholder.
status: open
firmness: default
semantic_text: The demo's model layer carries D-208's spend guards, as the core's flags (a positive daily budget, 10 questions an hour and 30 a day per session and per address, 3 calls at once, the trusted proxy of the address header and Turnstile's place in the policy) and the built-in UI's config and grant (the limits of one question, a site key, the verification host as an egress host and the Secret ui--turnstile-secret), checked by validate.sh, which also fails a live demo that includes the layer while the site key is the placeholder.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-15T00:17:05.244Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3975336e8831fb3944623d74ceefc23c073dc681316e419f19fd587405d680c8
semantic_inventory:
  - claim_key: CLAIM-6C208CBDBEA8E9E2
    claim_text: The demo's model layer carries D-208's spend guards, as the core's flags (a positive daily budget, 10 questions an hour and 30 a day per session and per address, 3 calls at once, the trusted proxy of the address header and Turnstile's place in the policy) and the built-in UI's config and grant (the limits of one question, a site key, the verification host as an egress host and the Secret ui--turnstile-secret), checked by validate.sh, which also fails a live demo that includes the layer while the site key is the placeholder
    role: descriptive
    status: missing
    span:
      start: 0
      end: 528
logic_claims:
  - CLAIM-6C208CBDBEA8E9E2
id: REQ-demo-model-guards-prepared
type: req
---
The demo's model layer carries D-208's spend guards, as the core's flags (a positive daily budget, 10 questions an hour and 30 a day per session and per address, 3 calls at once, the trusted proxy of the address header and Turnstile's place in the policy) and the built-in UI's config and grant (the limits of one question, a site key, the verification host as an egress host and the Secret ui--turnstile-secret), checked by validate.sh, which also fails a live demo that includes the layer while the site key is the placeholder.

## Context

D-208 guards the demo's model spend with the provider's hard limit, a daily budget, limits per session and per address, calls at once and Turnstile. The core's flags are off by default, so only the demo's model layer turns them on, and the layer stays out of the live demo until the owner includes it. validate.sh checks the built layer so a guard cannot be dropped or changed by accident, and refuses a live demo that includes the layer while the Turnstile site key is the placeholder the author replaces after creating the widget.

## Source

> The author's M5 brief: the demo's spend guards (D-208) prepared in the model layer, off until the model is turned on (M5 brief from the design session, 2026-11-14).
---
title: 'The demo''s model layer is included in the live demo, with the switch still off until the owner turns it on, an audited owner action: an OpenRouter provider instance with its key in a named Secret and egress to openrouter.ai only, the model entry, the Ask box''s route and the tier lists, stored by the demo loader only when the database has none and never turning the switch or a toggle on; the maintainer has no route; and the live demo makes no model call until the owner turns the model on.'
status: open
firmness: default
semantic_text: 'The demo''s model layer is included in the live demo, with the switch still off until the owner turns it on, an audited owner action: an OpenRouter provider instance with its key in a named Secret and egress to openrouter.ai only, the model entry, the Ask box''s route and the tier lists, stored by the demo loader only when the database has none and never turning the switch or a toggle on; the maintainer has no route; and the live demo makes no model call until the owner turns the model on.'
origin:
  kind: agent
  ref: 'The author''s decision of 2026-11-17 (the brief for the change that includes the demo''s model layer in the live demo: key Secret there, Turnstile secret there, daily budget 0.25, address check works normally; the switch stays the owner''s), on D-203, D-207, D-208, D-221'
  recorded_at: '2026-11-17T21:46:41.810Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9fa52587440f74b0e42f843f689260b5af94b7b0ccb284d845edbe8998a15482
semantic_inventory:
  - claim_key: CLAIM-DFA1DC28D1F5357E
    claim_text: 'The demo''s model layer is included in the live demo, with the switch still off until the owner turns it on, an audited owner action: an OpenRouter provider instance with its key in a named Secret and egress to openrouter.ai only, the model entry, the Ask box''s route and the tier lists, stored by the demo loader only when the database has none and never turning the switch or a toggle on'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 388
  - claim_key: CLAIM-150C71D09CCBB836
    claim_text: the maintainer has no route
    role: descriptive
    status: missing
    span:
      start: 390
      end: 417
  - claim_key: CLAIM-4E63F57755455EA8
    claim_text: and the live demo makes no model call until the owner turns the model on
    role: descriptive
    status: missing
    span:
      start: 419
      end: 491
logic_claims:
  - CLAIM-DFA1DC28D1F5357E
  - CLAIM-150C71D09CCBB836
  - CLAIM-4E63F57755455EA8
id: REQ-demo-model-prepared-off
type: req
---
The demo's model layer is included in the live demo, with the switch still off until the owner turns it on, an audited owner action: an OpenRouter provider instance with its key in a named Secret and egress to openrouter.ai only, the model entry, the Ask box's route and the tier lists, stored by the demo loader only when the database has none and never turning the switch or a toggle on; the maintainer has no route; and the live demo makes no model call until the owner turns the model on.

## Context

D-207 picks the demo's model and D-208 guards its spend. The author has now decided to include the layer in the live demo (OpenRouter key Secret is there, Turnstile secret is there, daily budget 0.25, the address check works normally), so the kustomize Component deploy/demo/model is listed by deploy/demo/openfief and built and checked there by `task demo:validate`; this replaces the earlier arrangement in which the live demo did not list it. The switch and the Ask box's toggle remain the owner's (D-212): turning them on in Settings > Models is an audited owner action that no seed or session performs, so the layer included makes no model call. The demo-seed Job stores the settings file through an environment variable. The route is one member, openai/gpt-oss-120b pinned to deepinfra/bf16, which passed the bench's runs 16-2 and 17 under scorer version 3 (D-207, D-221); its provider instance states retention of 0 days and no training from DeepInfra's Terms section 7(b) (R-33). There is no fallback model, since no second host is a D-221 member yet: if DeepInfra is down the Ask box still answers its fixed commands (D-203) and says the model is unavailable for the rest.

## Source

> The author's M4 brief: prepare everything so the demo can use a model by flipping configuration, but keep it off (no model call) until the author sets the quality bar; switch off and Ask box toggle off, the maintainer with no route (M4 prep brief from the design session, 2026-11-14).

> The author's decision, given to the session that made the change: the OpenRouter key Secret "is there", the Turnstile secret is there, the daily budget is "0.25", and the address check "works normally"; step 1 of docs/dev/demo.md "Turning the model on" is the reviewed change that includes the model layer in the live demo, and step 3, the switch, stays the owner's audited action (2026-11-17).

---
title: 'The demo''s model is prepared as a layer the live demo does not include: an OpenRouter provider instance with its key in a named Secret and egress to openrouter.ai only, the model entry, the Ask box''s route and the tier lists, stored by the demo loader only when the database has none and never turning the switch or a toggle on; the maintainer has no route; and the live demo makes no model call until the owner turns the model on.'
status: open
firmness: default
semantic_text: 'The demo''s model is prepared as a layer the live demo does not include: an OpenRouter provider instance with its key in a named Secret and egress to openrouter.ai only, the model entry, the Ask box''s route and the tier lists, stored by the demo loader only when the database has none and never turning the switch or a toggle on; the maintainer has no route; and the live demo makes no model call until the owner turns the model on.'
origin:
  kind: agent
  ref: 'M4 prep brief from the design session: the demo''s model behind the switch (D-207 amended, D-205, D-208, D-212, D-203), 2026-11-14'
  recorded_at: '2026-11-14T17:59:15.660Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8d536c4c088d44acbbe700674d72c3802f1112b52fca980910d3631c985ee27b
semantic_inventory:
  - claim_key: CLAIM-F36F8B50F9FC3D31
    claim_text: 'The demo''s model is prepared as a layer the live demo does not include: an OpenRouter provider instance with its key in a named Secret and egress to openrouter.ai only, the model entry, the Ask box''s route and the tier lists, stored by the demo loader only when the database has none and never turning the switch or a toggle on'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 327
  - claim_key: CLAIM-150C71D09CCBB836
    claim_text: the maintainer has no route
    role: descriptive
    status: missing
    span:
      start: 329
      end: 356
  - claim_key: CLAIM-4E63F57755455EA8
    claim_text: and the live demo makes no model call until the owner turns the model on
    role: descriptive
    status: missing
    span:
      start: 358
      end: 430
logic_claims:
  - CLAIM-F36F8B50F9FC3D31
  - CLAIM-150C71D09CCBB836
  - CLAIM-4E63F57755455EA8
id: REQ-demo-model-prepared-off
type: req
---
The demo's model is prepared as a layer the live demo does not include: an OpenRouter provider instance with its key in a named Secret and egress to openrouter.ai only, the model entry, the Ask box's route and the tier lists, stored by the demo loader only when the database has none and never turning the switch or a toggle on; the maintainer has no route; and the live demo makes no model call until the owner turns the model on.

## Context

D-207 picks the demo's model and D-208 guards its spend, but the author has not set the quality bar, so nothing may reach OpenRouter yet. The switch and the Ask box's toggle are the owner's (D-212) and the key is created out of band, so what is prepared is a kustomize Component that deploy/demo/openfief does not list, built and checked by `task demo:validate`, and a settings file the demo-seed Job stores through an environment variable. The fallback (openai/gpt-oss-120b on one host) waits for the next bench run to name the host: a placeholder host would pin to nothing, so the entry is neither in the plugin nor in the route until then.

## Source

> The author's M4 brief: prepare everything so the demo can use a model by flipping configuration, but keep it off (no model call) until the author sets the quality bar; nothing may change the live demo's behaviour now; switch off and Ask box toggle off, the maintainer with no route (M4 prep brief from the design session, 2026-11-14).

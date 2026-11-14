---
title: The ModelSettingsService must list the model-provider instances that run, with where each declares its models run, its retention terms and its models, to the owner alone, an instance that does not answer as unavailable, and must name the entries of every route that name a missing provider instance or a plugin that is no model provider.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the MS1 brief, from D-205 and D-196, 2026-11-13
  recorded_at: '2026-11-14T00:24:19.688Z'
semantic_text: The ModelSettingsService must list the model-provider instances that run, with where each declares its models run, its retention terms and its models, to the owner alone, an instance that does not answer as unavailable, and must name the entries of every route that name a missing provider instance or a plugin that is no model provider.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 45899246acd32daaa77e89975990ade948813590dcb7797926719141bbaae778
semantic_inventory:
  - claim_key: CLAIM-9350F7209D7250D2
    claim_text: The ModelSettingsService must list the model-provider instances that run, with where each declares its models run, its retention terms
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 134
  - claim_key: CLAIM-6594515D6B56A580
    claim_text: its models, to the owner alone, an instance that does not answer as unavailable, and must name the entries of every route that name a missing provider instance or a plugin that is no model provider
    role: normative
    status: ontology_gap
    span:
      start: 139
      end: 336
logic_claims:
  - CLAIM-9350F7209D7250D2
  - CLAIM-6594515D6B56A580
semantic_clauses:
  - The ModelSettingsService must list the model-provider instances that run, with where each declares its models run, its retention terms
  - its models, to the owner alone, an instance that does not answer as unavailable, and must name the entries of every route that name a missing provider instance or a plugin that is no model provider
id: REQ-models-providers-listed-to-owner
type: req
---
The ModelSettingsService must list the model-provider instances that run, with where each declares its models run, its retention terms and its models, to the owner alone, an instance that does not answer as unavailable, and must name the entries of every route that name a missing provider instance or a plugin that is no model provider.

## Context

The owner chooses routes and tier lists from what the plugins list, so the core hands the Settings page the providers and the broken entries, routes in use or not.

## Source

> Providers: list model-provider plugin instances with their models (id, prices, limits, tool/structured flags) and each provider's declared location and retention terms shown as information (D-205).

Source: the MS1 brief, from D-205 and D-196, 2026-11-13

---
title: A plugin image must be refused unless its config's io.openfief.plugin.manifest label parses strictly as a valid manifest equal, as sets, to the manifest its Plugin declares.
status: open
text_ref: docs/DESIGN.md#D-42
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-42
  - lane:strict
semantic_text: A plugin image must be refused unless its config's io.openfief.plugin.manifest label parses strictly as a valid manifest equal, as sets, to the manifest its Plugin declares.
logic_claims:
  - CLAIM-D92AECF23AC49182
semantic_clauses:
  - A plugin image must be refused unless its config's io.openfief.plugin.manifest label parses strictly as a valid manifest equal, as sets, to the manifest its Plugin declares.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9bb435f7aebd7f273e08e968cc91d1af436800878c62e16d0673606a02c5ea89
semantic_inventory:
  - claim_key: CLAIM-D92AECF23AC49182
    claim_text: A plugin image must be refused unless its config's io.openfief.plugin.manifest label parses strictly as a valid manifest equal, as sets, to the manifest its Plugin declares
    role: exception
    status: modeled
    span:
      start: 0
      end: 172
origin:
  kind: agent
  recorded_at: '2026-10-28T08:52:04.479Z'
id: REQ-plugins-manifest-label-match
type: req
---
A plugin image must be refused unless its config's io.openfief.plugin.manifest label parses strictly as a valid manifest equal, as sets, to the manifest its Plugin declares.

## Context

D-42 embeds the manifest in the image as a config label so one signature covers the code and the permissions it asks for. The Plugin resource still declares spec.manifest (the CLI copies it from the label, D-46), so the operator checks the two are the same; a Plugin that declares more than the signed label is refused rather than trusted. Unknown label fields are refused so a newer format is never half-read. Refines REQ-plugins-manifest-in-image. Modeled by Claude Code for the M0 signature-verification task; the author has not reviewed it yet.

## Source

> A plugin's manifest is embedded in its image, as a label on the image config. One signature covers both the code and the permissions it asks for, and the manifest can be read from the registry without running the image.

docs/DESIGN.md D-42

---
title: Each first-party plugin's image must carry its committed manifest.json, compacted, as the io.openfief.plugin.manifest label, and task check must fail when a Dockerfile's label or a demo Plugin's spec.manifest differs from it.
status: open
text_ref: docs/DESIGN.md#D-42
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-42
  - lane:strict
semantic_text: Each first-party plugin's image must carry its committed manifest.json, compacted, as the io.openfief.plugin.manifest label, and task check must fail when a Dockerfile's label or a demo Plugin's spec.manifest differs from it.
logic_claims:
  - CLAIM-CF8786E4DAD30C57
semantic_clauses:
  - Each first-party plugin's image must carry its committed manifest.json, compacted, as the io.openfief.plugin.manifest label, and task check must fail when a Dockerfile's label or a demo Plugin's spec.manifest differs from it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f57a398ecd80823914bdc994eef2ce8c59cb150bb4cf936007653dc8206a3b3b
semantic_inventory:
  - claim_key: CLAIM-CF8786E4DAD30C57
    claim_text: Each first-party plugin's image must carry its committed manifest.json, compacted, as the io.openfief.plugin.manifest label, and task check must fail when a Dockerfile's label or a demo Plugin's spec.manifest differs from it
    role: normative
    status: modeled
    span:
      start: 0
      end: 224
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:40.011Z'
id: REQ-plugins-first-party-manifest-label
type: req
---
Each first-party plugin's image must carry its committed manifest.json, compacted, as the io.openfief.plugin.manifest label, and task check must fail when a Dockerfile's label or a demo Plugin's spec.manifest differs from it.

## Context

D-42 embeds the manifest in the image as a label covered by the signature, and the operator refuses a Plugin whose spec.manifest differs from the label (REQ-plugins-manifest-label-match). Keeping one committed source and a check in task check means a permission change can't reach an image without review, and the demo's Plugin can't drift from the image it deploys. Refines REQ-plugins-manifest-in-image. Modeled by Claude Code for the M0 plugin binaries task (agent A: SDK verifier, operator key publishing, first-party plugins); the author has not reviewed it yet.

## Source

> A plugin's manifest is embedded in its image, as a label on the image config. One signature covers both the code and the permissions it asks for, and the manifest can be read from the registry without running the image.

docs/DESIGN.md D-42

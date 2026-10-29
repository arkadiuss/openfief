---
title: Keeps Shared Secrets = true
status: active
text_ref: docs/DESIGN.md#D-58
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-58
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.uninstall_cli
property_key: keeps_shared_secrets
canonical_key: mcp-kibi-model-requirement:plugins.uninstall_cli:keeps_shared_secrets:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1CF1C25F56D4B1F1
claim_text: Uninstalling an instance must delete the Secrets its manifest or grant names, except those another installed instance still names, which it must keep and report.
origin:
  kind: agent
  recorded_at: '2026-10-28T22:24:14.561Z'
id: FACT-PROP-8C24FB3C400B8510
type: fact
---

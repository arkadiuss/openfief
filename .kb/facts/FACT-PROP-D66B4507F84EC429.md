---
title: First Party Entries On Release Commit = true
status: active
text_ref: docs/DESIGN.md#D-88
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-88
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.release_catalog
property_key: first_party_entries_on_release_commit
canonical_key: mcp-kibi-model-requirement:tooling.release_catalog:first_party_entries_on_release_commit:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4E2FDA7909D37F66
claim_text: A published release must add each plugins/<name> image to the catalog with the digest it signed, the release version and OpenFief's CI as publisher, refusing a missing digest or a version already listed, on a release commit from which the operator, core and CLI of that release are built.
origin:
  kind: agent
  recorded_at: '2026-10-29T18:02:06.242Z'
id: FACT-PROP-D66B4507F84EC429
type: fact
---

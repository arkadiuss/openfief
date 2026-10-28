---
title: Manages Only Docs Record = true
status: active
text_ref: docs/dev/demo.md#dns
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-dns
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: docs.dns
property_key: manages_only_docs_record
canonical_key: mcp-kibi-model-requirement:docs.dns:manages_only_docs_record:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-044F396F49BFC390
claim_text: task dns:docs must make docs.<DEMO_DOMAIN> a DNS-only CNAME to the GitHub Pages host with the comment managed-by=openfief-docs, creating or updating only a record with that comment, refusing when any other record occupies the name, and task dns:docs:plan must show the change without making it; the Cloudflare token is handled as for the demo.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:36:17.533Z'
id: FACT-PROP-04561E496879DDFF
type: fact
---

---
title: Manages Only Commented Records = true
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
subject_key: demo.dns
property_key: manages_only_commented_records
canonical_key: mcp-kibi-model-requirement:demo.dns:manages_only_commented_records:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8AE21BA437332713
claim_text: With CLOUDFLARE_API_TOKEN set and DEMO_DOMAIN not none, the demo automation must point the demo's host name <ENV>.<DEMO_DOMAIN> (demo.openfief.dev for the demo named demo) at its server with a DNS-only A record and a CAA record that allows only letsencrypt.org, each with the comment managed-by=openfief-demo env=<env>, creating missing records and updating changed ones in place, keeping them across recreate, and deleting them on down; it must never update or delete a record without that comment, and must refuse when such a record occupies the host name.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:35:43.897Z'
id: FACT-PROP-B6F663658FA4B986
type: fact
---

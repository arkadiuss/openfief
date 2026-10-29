---
title: Manages Only Site Records = true
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
subject_key: site.dns
property_key: manages_only_site_records
canonical_key: mcp-kibi-model-requirement:site.dns:manages_only_site_records:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A05EEF8217267293
claim_text: task dns:site must point <DEMO_DOMAIN> at GitHub Pages with DNS-only A and AAAA records to exactly GitHub's published Pages addresses and www.<DEMO_DOMAIN> with a DNS-only CNAME to <owner>.github.io, and with GITHUB_PAGES_CHALLENGE set a TXT record _github-pages-challenge-<owner>.<DEMO_DOMAIN>, each with the comment managed-by=openfief-site; it must create, update or delete only records with that comment, delete records with the retired comment managed-by=openfief-docs, never create a docs record, refuse before any change when another record of the same name and type or a conflicting CNAME exists, and task dns:site:plan must show the changes without making them; the Cloudflare token is handled as for the demo.
origin:
  kind: agent
  recorded_at: '2026-10-28T15:29:33.086Z'
id: FACT-PROP-485D8FCCA88F0623
type: fact
---

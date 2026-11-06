---
title: Catalog Pinned By Test = fail
status: active
text_ref: docs/dev/code-organization.md#architecture-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-architecture-tests
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.archtest
property_key: catalog_pinned_by_test
canonical_key: docs-dev-code-organization-md:tooling.archtest:catalog_pinned_by_test:eq:fail
operator: eq
value_type: string
value_string: fail
claim_key: CLAIM-403971453AC7E999
claim_text: 'The architecture tests must fail when a test pins the embedded plugin catalog: it compares the number of catalog.Load() entries with a literal, picks an entry by a literal index, spells out a digest the catalog pins, or inspects what a catalog-reading function of its package returns without deriving its expectation from catalog.Load().'
origin:
  kind: agent
  recorded_at: '2026-11-06T17:50:02.114Z'
id: FACT-PROP-0AB6953BE08F32E8
type: fact
---

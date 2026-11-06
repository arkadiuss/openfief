---
title: Foreign Schema In Sql = fail
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
property_key: foreign_schema_in_sql
canonical_key: docs-dev-code-organization-md:tooling.archtest:foreign_schema_in_sql:eq:fail
operator: eq
value_type: string
value_string: fail
claim_key: CLAIM-14088189B2AC0C8B
claim_text: The architecture tests must fail when a context's SQL, in its migrations or in its packages' query strings, names a schema that another context owns.
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:49.029Z'
id: FACT-PROP-BC05B9683FAA33D9
type: fact
---

---
title: Commit Without Author Signoff = fail
status: active
text_ref: design-decisions:D-5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-5
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.dco_check
property_key: commit_without_author_signoff
canonical_key: docs-design-md:tooling.dco_check:commit_without_author_signoff:eq:fail
operator: eq
value_type: string
value_string: fail
claim_key: CLAIM-D9E99063699F8A78
claim_text: The DCO check must fail when a commit in a pull request has no Signed-off-by trailer naming its author.
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:03.138Z'
id: FACT-PROP-C129B881A51EE7F5
type: fact
---

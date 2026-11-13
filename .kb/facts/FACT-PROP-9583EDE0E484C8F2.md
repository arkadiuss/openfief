---
title: Staged Check Saves Throwaway Store = false
status: active
text_ref: docs/dev/kibi.md#setup-notes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-setup-notes
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.kibi_gate
property_key: staged_check_saves_throwaway_store
canonical_key: docs-dev-kibi-md-setup-notes:tooling.kibi_gate:staged_check_saves_throwaway_store:bool:false
operator: eq
value_type: bool
value_bool: false
claim_key: CLAIM-30DFEB40D9896388
claim_text: The pre-commit kibi step must keep the throwaway knowledge base that kibi check --staged builds off the disk while it projects the branch's records into it, and must remove the throwaway knowledge bases that killed runs left in the temporary directory before it runs the checks, so that a one-file commit's staged check does not grow with the square of the knowledge base.
origin:
  kind: agent
  recorded_at: '2026-11-13T02:54:34.302Z'
id: FACT-PROP-9583EDE0E484C8F2
type: fact
---

---
title: Cli Runs Core Enrollment Through Exec With Owner Kubeconfig = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-80-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-80#1
fact_kind: property_value
subject_key: identity.passkey_enrollment
property_key: cli_runs_core_enrollment_through_exec_with_owner_kubeconfig
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:identity.passkey_enrollment:cli_runs_core_enrollment_through_exec_with_owner_kubeconfig:eq:true
claim_key: CLAIM-50D4CBD2D093CBF7
claim_text: The openfief CLI must have a command for the first passkey that runs the core's enrollment command in the core's pod through Kubernetes exec, with the owner's kubeconfig.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:09:42.425Z'
id: FACT-PROP-D2C4269644DB050E
type: fact
---

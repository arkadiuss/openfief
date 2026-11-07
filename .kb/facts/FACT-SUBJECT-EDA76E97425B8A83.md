---
title: Plugins Ca Bundle Cli
status: active
text_ref: design-session:connector-endpoint-ca-bundle-cli
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-bundle-cli
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.ca_bundle_cli
canonical_key: plugins.ca_bundle_cli
claim_key: CLAIM-58C06F6FE963BCF7
claim_text: openfief plugin ca-bundle set <instance> <pem-file> must accept only a file of PEM certificates, then create or rotate the Secret <instance>--ca-bundle (key ca.crt, labelled as the instance) and stamp the SHA-256 of what it holds on the Plugin spec, and with --print must emit both for GitOps with a note that the owner keeps the checksum in step with the Secret.
origin:
  kind: agent
  recorded_at: '2026-11-07T17:33:03.367Z'
id: FACT-SUBJECT-EDA76E97425B8A83
type: fact
---

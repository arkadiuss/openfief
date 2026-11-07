---
title: Sets Secret And Stamps Checksum = true
status: active
text_ref: design-session:connector-endpoint-ca-bundle-cli
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-bundle-cli
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.ca_bundle_cli
property_key: sets_secret_and_stamps_checksum
canonical_key: mcp-kibi-model-requirement:plugins.ca_bundle_cli:sets_secret_and_stamps_checksum:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-58C06F6FE963BCF7
claim_text: openfief plugin ca-bundle set <instance> <pem-file> must accept only a file of PEM certificates, then create or rotate the Secret <instance>--ca-bundle (key ca.crt, labelled as the instance) and stamp the SHA-256 of what it holds on the Plugin spec, and with --print must emit both for GitOps with a note that the owner keeps the checksum in step with the Secret.
origin:
  kind: agent
  recorded_at: '2026-11-07T17:33:04.144Z'
id: FACT-PROP-7A0DEF3C5A02EC4E
type: fact
---

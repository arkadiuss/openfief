---
title: Sdk Go Plugin Environment
status: active
text_ref: docs/DESIGN.md#D-11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-11
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: sdk.go.plugin_environment
canonical_key: sdk.go.plugin_environment
claim_key: CLAIM-B65D6BBFA0BD7609
claim_text: 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.'
origin:
  kind: agent
  recorded_at: '2026-10-28T08:02:23.489Z'
id: FACT-SUBJECT-BB036C8C97045979
type: fact
---

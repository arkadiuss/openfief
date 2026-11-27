---
title: 'While a core under SpiceDB is not ready, its /readyz answer and its log must name what is missing: the Secret openfief-spicedb-key and the file it is read from while the key is absent, the role or the database spicedb that the openfief-db cluster lacks, a preshared key SpiceDB refuses, or SpiceDB not answering at its address.'
status: open
firmness: default
required_evidence: unverified
semantic_text: 'While a core under SpiceDB is not ready, its /readyz answer and its log must name what is missing: the Secret openfief-spicedb-key and the file it is read from while the key is absent, the role or the database spicedb that the openfief-db cluster lacks, a preshared key SpiceDB refuses, or SpiceDB not answering at its address.'
rationale: 'The design session, relaying the author''s decision that SpiceDB becomes the default authorizer before 2026.11.0-alpha.2 (D-164), asked for it: an owner who upgrades without SpiceDB''s role and database, or without its Secret, must read what is missing instead of a silent not-ready.'
origin:
  kind: agent
  ref: 'the design session''s task for the SpiceDB default (D-164), 2026-11-22: Loud failure mode. On an upgrade where the owner skipped install step 5 (the database file''s SpiceDB Database and managed role), or SpiceDB''s Secret is missing, the core must refuse every authorization check, stay not ready, and say exactly what''s missing in its readiness output and logs: the Secret by name, or the role/database not yet reconciled. A silent not-ready is not acceptable.'
  recorded_at: '2026-11-22T08:39:29.852Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: be212404a9ee24b7054aff707edbc105f99d3278fcd78f58629588f4c4f758cc
semantic_inventory:
  - claim_key: CLAIM-F3205070A0C10850
    claim_text: 'While a core under SpiceDB is not ready, its /readyz answer and its log must name what is missing: the Secret openfief-spicedb-key and the file it is read from while the key is absent, the role or the database spicedb that the openfief-db cluster lacks, a preshared key SpiceDB refuses, or SpiceDB not answering at its address'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 326
    subjects:
      - core.spicedb_not_ready_reason
logic_claims:
  - CLAIM-F3205070A0C10850
id: REQ-core-spicedb-not-ready-says-why
type: req
---
While a core under SpiceDB is not ready, its /readyz answer and its log must name what is missing: the Secret openfief-spicedb-key and the file it is read from while the key is absent, the role or the database spicedb that the openfief-db cluster lacks, a preshared key SpiceDB refuses, or SpiceDB not answering at its address.

## Context

The design session asked for this on 2026-11-22 while relaying the author's decision that SpiceDB answers authorization by default from 2026.11.0-alpha.2 (D-164, D-163 unchanged): the release bundle and the demo run the core with --authz=spicedb, and an owner upgrading from 2026.11.0-alpha.1 who skipped install step 5, or whose Secret is missing, gets a core that refuses every check and stays not ready. The readiness gate itself is the start's alone, as REQ-identity-authz-spicedb-answers and SCEN-core-spicedb-waits-for-key say; this requirement is about what it says. The core asks its own database's catalog (pg_roles, pg_database), which every role may read, whether SpiceDB's role and database exist, reads the failed call's gRPC code, and never quotes an error. Built in cmd/core/authzwait.go; the key gate's warning repeats every minute while the key is missing.

## Source

> Loud failure mode. On an upgrade where the owner skipped install step 5 (the database file's SpiceDB Database and managed role), or SpiceDB's Secret is missing, the core must refuse every authorization check, stay not ready, and say exactly what's missing in its readiness output and logs: the Secret by name, or the role/database not yet reconciled. A silent not-ready is not acceptable.

Source: the design session's task for the SpiceDB default, 2026-11-22.

---
title: The core writes the embedded schema to SpiceDB before it reconciles, at every start
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:16.886Z'
id: SCEN-identity-spicedb-schema-written-at-start
type: scenario
---
Given the core started with --authz=spicedb and a SpiceDB that holds no schema, when it reconciles, then the schema of authz/schema.zed, embedded in the binary, is written before any relationship is read or written, and writing the schema SpiceDB already holds changes nothing; a SpiceDB that refuses the schema stops the reconcile with no relationship written, and the core retries.
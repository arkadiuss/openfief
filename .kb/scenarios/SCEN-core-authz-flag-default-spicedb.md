---
title: With no --authz flag the core answers from SpiceDB; --authz=policy answers from the grants table and needs nothing of SpiceDB
status: active
origin:
  kind: agent
  recorded_at: '2026-11-22T08:42:38.175Z'
id: SCEN-core-authz-flag-default-spicedb
type: scenario
---
Given the core started with no --authz flag, as serve or as core token issue or revoke, then it runs SpiceDB (D-164): it reads SpiceDB's endpoint from --spicedb-endpoint or OPENFIEF_SPICEDB_ENDPOINT, plaintext gRPC from --spicedb-insecure or OPENFIEF_SPICEDB_INSECURE=true, and the preshared key from the environment or its file, and without an endpoint it exits at once with an error that names both and the opt-out --authz=policy instead of waiting for a SpiceDB that isn't there. Given the core started with --authz=policy, with or without SpiceDB's settings in the environment, then the authorizer is the policy over the grants table, no connection to SpiceDB is made and the registry has no observer. Replaces SCEN-core-authz-flag-default-policy, written while the policy was the default. Assumes the core's Deployment sets SpiceDB's endpoint and plaintext setting in the environment, as deploy/core does.
---
title: core init-credentials creates SpiceDB's preshared key and its role's password as two Secrets of 32 random bytes each, and nothing else
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T07:41:52.485Z'
id: SCEN-identity-credentials-created-when-missing
type: scenario
---
Given the openfief namespace holds neither Secret, when core init-credentials runs with the ServiceAccount of the credentials Job, then it creates openfief-spicedb-key (one key) and openfief-db-spicedb (a kubernetes.io/basic-auth Secret for the role spicedb, labelled for CloudNativePG's reload), each from 32 bytes of the system's random source as 64 hex digits, different from each other; the API calls are two creates of Secrets in openfief and no other call; and the log names each Secret as created and holds neither value.
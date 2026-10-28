---
title: The generated AuthService client reaches the generated handler over Connect
status: active
sourceFile: internal/identity/connectapi/contract_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T18:34:52.652Z'
id: TEST-identity-auth-contract
type: test
---
Mounts the generated AuthService handler, with the generated unimplemented handler behind it, on an httptest server, checks that it is served at /openfief.identity.v1.AuthService/, and calls BeginLogin through the generated Connect client with a synthetic localhost origin. The call must reach the handler and come back as a Connect unimplemented error, which proves the client, the handler, the service path and the messages agree. The core's real handler is tested separately once it exists.
---
title: The SpiceDB authorizer asks SpiceDB for the account permission and fails closed
status: active
sourceFile: internal/identity/spicedb/authorizer_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:24.705Z'
id: TEST-identity-spicedb-authorizer
type: test
---
Over a fake of SpiceDB's permissions service: each holder kind (owner as the one person, script token, view plugin, demo session) is checked as the account's view or change permission of its own subject, fully consistent, with IDs SpiceDB can't take escaped; only HAS_PERMISSION allows; for every holder and action a failing SpiceDB and a timeout answer false with the error, the owner included; an empty ID is refused without a call; the list answer is one LookupResources call, decoded, with conditional results left out, and no accounts at all on any failure. A false pass would be a fake that answers for the owner on error, which the owner rows of the fail-closed table would catch. A failing, slow or pinged-in-vain SpiceDB is reported as the unavailable error, an empty ID and a client that went away are not.
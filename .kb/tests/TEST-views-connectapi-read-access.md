---
title: The read API serves D-137's read access and nothing more
status: active
sourceFile: internal/views/connectapi/readaccess_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T10:05:26.954Z'
id: TEST-views-connectapi-read-access
type: test
---
TestReadAccessMethods enumerates every method of the ReadService descriptor, fails on one not placed against D-137's list, checks every kind of record D-137 names is served, and that each method declares no side effects and is a GET without a body. TestReadAccessReturnsNoSecrets walks every message a served method returns for a field named like a secret, credential, passkey, session or token or a broker account. TestReadAccessAttentionNamesNoBrokerAccount checks over REST that an unmapped broker account's attention item carries no broker account number (D-121, D-137).
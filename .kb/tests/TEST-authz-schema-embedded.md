---
title: The embedded schema is authz/schema.zed as it stands
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:15.997Z'
sourceFile: authz/authz_test.go
verification_scope: unit
id: TEST-authz-schema-embedded
type: test
---
Reads authz/schema.zed and compares it with authz.Schema, the copy the core writes to SpiceDB, and checks it holds the household, sleeve and account definitions. A false pass would be an embedded file that is another one, which the byte comparison rules out.
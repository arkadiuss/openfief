---
title: Through the real core and PostgreSQL, a sync re-keys a lot booked under the derived ISIN, and later syncs change nothing
status: active
sourceFile: cmd/core/rekey_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:23.367Z'
id: TEST-core-rekey
type: test
---
Seeds a COI lot under the ISIN earlier versions derived from its series code with its terms, the owner's exemption class, a delivered display name and a model target, binds a fake treasury-bonds connector that books coupons and answers registry ISINs, and syncs three times: the first names the instrument as provisional and re-keys everything, the next two name nothing and leave the coupons unchanged; every current transaction posts to the registry's instrument under its old key, and the terms, class, name and model target moved (task test:pg).
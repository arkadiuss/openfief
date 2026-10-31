---
title: The owner deletes a sleeve with no accounts and a kind no sleeve has
status: active
sourceFile: internal/portfolio/app/deletions_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:06.930Z'
id: TEST-portfolio-app-deletions
type: test
---
Steps through portfolio/app.Sleeves.DeleteSleeve and DeleteKind with a fake store: a sleeve with accounts and a kind in use are refused; deleting the satellite sleeve frees its kind, which is then deleted; unknown ones are NotFound; a read session is refused; the bonds sleeve stays.
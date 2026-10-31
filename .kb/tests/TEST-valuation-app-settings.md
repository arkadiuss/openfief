---
title: 'The owner''s settings: the display settings for anyone, every setting and its changes for the owner only'
status: active
sourceFile: internal/valuation/app/settings_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:29.994Z'
id: TEST-valuation-app-settings
type: test
verification_scope: unit
---
Use-case tests over fakes: the display settings give PLN and the pack's default number format until the owner chooses one; the owner reads every setting with the pack's wrappers and the marginal rate; a read session is refused; the owner's number format and marginal rate are recorded, invalid ones refused storing nothing.
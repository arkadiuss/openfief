---
title: zed validates the authorization schema against its relationships and assertions
status: active
sourceFile: authz/validation.yaml
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T22:46:08.012Z'
id: TEST-identity-authz-schema
type: test
---
task authz:validate runs zed validate on authz/validation.yaml: synthetic relationships for one household and two accounts, and assertions that the owner views and changes every account, every reader views and never changes, a viewer views only its granted account, and a stranger views nothing. A failing assertion fails task check.
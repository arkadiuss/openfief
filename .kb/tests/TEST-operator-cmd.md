---
title: The operator binary's flags default to the openfief Services and refuse a bad plugin port
status: active
sourceFile: cmd/operator/main_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:13:05.176Z'
id: TEST-operator-cmd
type: test
verification_scope: unit
---
Table tests call parseFlags with no arguments, with every Service, port and plugin-port flag set, and with plugin ports 0 and 70000, and compare the resulting settings or the sentinel error with go-cmp. The manager wiring itself is exercised by TestSetupWithManager in the operator package. A false pass would be a default that names another namespace's Service; the defaults are compared in full.
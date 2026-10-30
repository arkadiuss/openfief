---
title: Sleeve kinds validate, keep their default role, and drive sleeve roles and install defaults
status: active
sourceFile: internal/portfolio/kind_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T09:34:18.315Z'
id: TEST-portfolio-sleeve-kind
type: test
---
Table tests of the portfolio package's kind rules: KindKey and Kind validation, the six built-in kinds with their default roles, CheckKind refusing a stored kind's role change, CheckSleeve filling a missing role and requiring the role override exactly when the role differs from the kind's default in either direction (D-71), the three default sleeves, MissingDefaults returning default sleeves only while none is stored, and KindBySleeve. A false pass would need the override check to ignore direction, which the defensive in an offensive kind case rules out.
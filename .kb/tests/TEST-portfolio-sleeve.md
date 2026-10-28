---
title: A sleeve has an ID, a name and exactly one role
status: active
sourceFile: internal/portfolio/sleeve_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:13:21.659Z'
id: TEST-portfolio-sleeve
type: test
---
Table tests in package portfolio_test validate synthetic sleeves through the public API: a defensive and an offensive sleeve pass; a sleeve with no role, an out-of-range role, no ID or no name fails with ErrInvalidSleeve, compared with cmpopts.EquateErrors. ParseRole accepts exactly the words defensive and offensive and round-trips them through String, and rejects capitalized, empty and placeholder text. A false pass would be accepting the zero role; it has its own case.
---
title: Each architecture rule on small fixture modules, and the baseline comparison
status: active
sourceFile: tools/archtest/rules_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:41.897Z'
id: TEST-tooling-archtest-rules
type: test
---
TestRules writes small txtar fixture modules (module example.com/fix) to a temporary directory, loads each with go/packages offline outside the workspace, and compares the exact violation keys each rule returns: an allowed and a forbidden context arrow, the kernel importing a context, an unknown context, the demo dataset; domain, app and adapter imports; plugins and the SDK importing internal/; banned package names, test-only ones included; foreign schemas in migrations and query strings, with comments ignored and river owned by valuation; floats in domain, app and contracts but not in adapters; goroutines and package state in the domain; money and raw errors in log calls, with LogSafe and IDs allowed; sleeps and skips through aliases and closures but not a parser's Skip; catalog tests pinning a size, an index or a digest, and a test inspecting a catalog reader's result without deriving from catalog.Load. TestPlaceOf, TestParseList and TestCompare check classification, list parsing and that fresh violations and stale lines are both reported. A false pass would be a rule returning extra or missing keys: the whole sorted key list is compared with go-cmp.

---
title: The repository passes every architecture rule against its baseline and exceptions
status: active
sourceFile: tools/archtest/repo_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:42.670Z'
id: TEST-tooling-archtest-repo
type: test
---
TestArchitecture loads this repository (both modules of the workspace, and the tests, SQL and contracts under internal, cmd, plugins, sdk, tools, test and proto), runs every rule with the embedded catalog's digests, and fails on each violation neither baseline.txt nor exceptions.txt lists and on each listed line that matches no violation, naming the rule, the key and the doc section. A false pass would need a rule that misses a break, which TestRules guards per rule.

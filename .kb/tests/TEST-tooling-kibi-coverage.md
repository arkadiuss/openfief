---
title: kibi-coverage.sh joins every page of the coverage report
status: active
sourceFile: scripts/tests/kibi-coverage.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T15:49:47.913Z'
id: TEST-tooling-kibi-coverage
type: test
---
Runs scripts/kibi-coverage.sh against a stand-in kibi serving a fixed list of requirements a page at a time: lists longer than, equal to a multiple of, shorter than one page and empty each come out whole and in order, and a missing argument fails.
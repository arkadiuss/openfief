---
title: A sleeping or skipping test fails the architecture test
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:57.441Z'
id: SCEN-tooling-archtest-sleep-skip
type: scenario
---
Given a test calling time.Sleep through an aliased import, a subtest calling Skipf on its *testing.T and a helper calling SkipNow on a testing.TB, when the stable-tests rule runs, then each is reported by file and function; a parser's own Skip method is not.

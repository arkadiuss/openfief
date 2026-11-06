---
title: A new break or a stale baseline line fails the architecture test
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T17:50:03.863Z'
id: SCEN-tooling-archtest-baseline-exact
type: scenario
---
Given violations found in the repository and the checked-in baseline and exceptions, when they are compared, then a violation no list names is reported as new and a line matching no violation is reported as stale with its line number; a malformed line, a line without a reason, an unknown rule or a duplicate line fails to parse.

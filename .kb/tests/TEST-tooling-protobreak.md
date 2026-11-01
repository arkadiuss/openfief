---
title: tools/protobreak excuses only the new exceptions an author decision names, and only in a commit marked breaking
status: active
sourceFile: tools/protobreak/judge_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T19:21:28.775Z'
id: TEST-tooling-protobreak
type: test
---
Table tests over a synthetic base contract (a Things service with ListGadgets and GetThing) and a synthetic DESIGN.md: judge passes a method removal with its new entries only when the decision names the method as code; it fails a removal without entries, any other breaking change beside an excused one, entries already on the base, an entry that excuses nothing, a message the removed method never held, and a decision that is missing or does not name the method. checkMarked passes an added entry only with `!` or a BREAKING CHANGE footer, and TestRun drives both subcommands over files as the script writes them. A false pass would need judge to accept a finding no new entry names, which every failing case would catch (D-157).
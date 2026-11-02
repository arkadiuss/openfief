---
title: A commit that adds a breaking-change exception is refused unless it is marked breaking
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T19:21:33.296Z'
id: SCEN-tooling-breaking-exception-commit-marked
type: scenario
---
Given a change that removes a method an author decision names and adds its entries to scripts/proto-breaking-exceptions.txt, when the commit message has neither `!` before the subject's colon nor a BREAKING CHANGE footer, then the commit-msg hook refuses it, and proto:breaking against the base refuses the range and names the commit; marked breaking, both pass, so the release notes list the change. Removing an entry needs no mark, and a merge commit is left to the range check.
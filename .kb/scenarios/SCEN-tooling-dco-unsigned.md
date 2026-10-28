---
title: An unsigned commit fails the DCO check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:21.051Z'
id: SCEN-tooling-dco-unsigned
type: scenario
---
Given a pull request whose range holds a commit with no Signed-off-by trailer, or one signed off by someone other than its author, when task dco runs on that range, then it exits non-zero and names the commit. When every commit is signed off by its author, it passes.

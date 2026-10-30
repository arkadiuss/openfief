---
title: A shared package change checks its importers only
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T03:32:46.321Z'
id: SCEN-tooling-hook-affected-packages
type: scenario
---
Given a staged change to one Go file of a package that other packages import, when the pre-commit hook runs, then lint and govulncheck run on that package and every package importing it directly or not, tests run on those and on every package whose tests import one of them, and no other Go package is checked; the hook prints each package with the reason it is included. Assumes go list describes the working tree.
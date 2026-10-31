---
title: The kibi gate recompiles once on a compile defect
status: active
sourceFile: scripts/tests/kibi-check.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-30T04:09:44.439Z'
id: TEST-tooling-kibi-check
type: test
verification_scope: integration
---
Runs scripts/kibi-check.sh with a fake kibi whose check answers come from a list: a valid KB passes with a single check; an authored violation fails at once with no recompile; a relationship missing from the compiled KB, or a duplicate row in the paginated entity projection, recompiles once with kibi sync --rebuild and the second check decides; a defect that survives the recompile fails with no third try. The call log is compared exactly, so an extra retry or a skipped recompile fails the test.
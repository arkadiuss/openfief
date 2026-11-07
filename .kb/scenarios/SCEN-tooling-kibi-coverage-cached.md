---
title: The docs build reuses the coverage report while its inputs are unchanged, never in CI
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T21:14:20.853Z'
id: SCEN-tooling-kibi-coverage-cached
type: scenario
---
Given a coverage report the docs build generated outside CI, when it builds again and the working-tree contents of the .kb files, the installed kibi-cli and kibi-core versions and both local kibi patches are unchanged, then it copies the kept report without calling kibi and prints that it reused the cached report for that key; when any of them changed, it asks kibi again and keeps the new report, with only the newest few kept. Under CI, with KIBI_COVERAGE_CACHE=off, or with proof runs ingested into .kb/proof/runs, it always asks kibi and neither reads nor writes the cache; a report kibi answered from a store that wasn't fresh is not kept.
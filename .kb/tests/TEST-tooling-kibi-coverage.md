---
title: kibi-coverage.sh joins every page of the coverage report and caches it outside CI
status: active
sourceFile: scripts/tests/kibi-coverage.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T21:14:19.316Z'
id: TEST-tooling-kibi-coverage
type: test
---
Runs scripts/kibi-coverage.sh in a throwaway Git repository against a stand-in kibi that serves a fixed list of requirements a page at a time and logs its calls. Paging: lists longer than, equal to a multiple of, shorter than one page and empty come out whole and in order, with one timed coverage query per page and no empty last page; the default page is 300 requirements, after one query that attaches the store; pages that repeat or miss requirements kibi counts fail and write no report; a missing argument is a usage error. Cache: a first run asks kibi and keeps the report; a second run with the same key doesn't call kibi and writes the same report; an ignored file under .kb doesn't change the key; a changed .kb file not yet staged, a new untracked .kb file, a changed kibi patch and another installed kibi-core version each regenerate, keeping only the three newest reports; CI=true, GITHUB_ACTIONS=true and KIBI_COVERAGE_CACHE=off regenerate although the report is cached and leave the cache alone, and so do ingested proof runs in .kb/proof/runs; a report from a store that isn't fresh is written but not kept.
---
title: Each page of the kibi coverage report is timed and warns the build from 90 seconds
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T19:27:54.684Z'
id: SCEN-tooling-kibi-slow-query-warned
type: scenario
---
Given the docs build asking kibi for the requirement coverage report 300 requirements at a time, after a cheap query has attached the store, when each page's command finishes, then its duration is printed; when it took 90 seconds or more, three quarters of kibi's 120-second query limit, a warning names the duration, as a workflow warning under GitHub Actions, while the page kibi wrote to stdout and kibi's exit status stay unchanged; the pages' rows are joined into one report that lists every requirement kibi counts exactly once, or the build fails.
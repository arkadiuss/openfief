---
title: A kibi coverage query from 60 seconds warns the build
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T19:27:54.684Z'
id: SCEN-tooling-kibi-slow-query-warned
type: scenario
---
Given the docs build asking kibi for the whole requirement coverage report in one query, when the command finishes, then its duration is printed; when it took 60 seconds or more, half of kibi's 120-second query limit, a warning names the duration, as a workflow warning under GitHub Actions, while the report written to stdout and kibi's exit status stay unchanged.
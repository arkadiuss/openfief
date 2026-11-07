---
title: The docs build reads requirement coverage a page at a time
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T15:49:48.588Z'
id: SCEN-docs-coverage-paged
type: scenario
---
Given a knowledge base whose whole coverage report takes kibi longer than its 120-second query limit, when the docs build generates the status pages, then it asks kibi for the report a page at a time and joins the pages, so every requirement is listed once and the build does not fail on the limit.
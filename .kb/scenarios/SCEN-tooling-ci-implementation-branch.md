---
title: A push to the implementation branch runs the full check and publishes nothing
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:48.473Z'
id: SCEN-tooling-ci-implementation-branch
type: scenario
---
Given a push to a branch named claude/openfief-implementation-*, when CI starts, then its check job runs task check with no condition, and the docs, image push and demo deploy jobs, which require github.ref to be main, do not run.

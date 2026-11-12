---
title: A bumped tool version builds a new cache
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:59.557Z'
id: SCEN-tooling-tool-cache
type: scenario
---
Given the setup action's cache step, when the Taskfile's pinned versions or go.mod's toolchain line change, then the cache key changes and the tools are installed once and cached again; the e2e job's state under .bin/e2e is never part of the cache.

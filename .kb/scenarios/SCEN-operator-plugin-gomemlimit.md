---
title: Every plugin container gets GOMEMLIMIT from its memory limit
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T18:23:48.795Z'
id: SCEN-operator-plugin-gomemlimit
type: scenario
---
Given a granted plugin whose container the operator limits to 256Mi of memory, when its Deployment is built, then the container's environment holds GOMEMLIMIT set to nine tenths of that limit in bytes, 241591905, and a container without a memory limit would get none.
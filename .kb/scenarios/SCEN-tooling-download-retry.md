---
title: A download is tried again after a 504 and not after a 404
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:53.390Z'
id: SCEN-tooling-download-retry
type: scenario
---
Given a download that answers 504 twice and then succeeds, when it runs through retry.sh, then it succeeds on the third try after waits of 2 and 4 seconds; given one that answers 404, it fails at once; given a server that never recovers, it gives up after six tries with waits up to 32 seconds.

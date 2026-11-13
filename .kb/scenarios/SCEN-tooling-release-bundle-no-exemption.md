---
title: A release bundle that sets the proxy's test-only exemption is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T23:20:24.984Z'
id: SCEN-tooling-release-bundle-no-exemption
type: scenario
---
Given the real deploy/ and a throwaway one whose egress proxy sets the exemption, once as a flag and once as an environment variable, when the bundle is rendered, then the real install holds neither, and each throwaway one fails with a message naming the exemption and writes no files.
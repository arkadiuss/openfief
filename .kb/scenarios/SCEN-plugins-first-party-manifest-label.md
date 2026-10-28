---
title: A first-party plugin's label always equals its manifest
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:00.625Z'
id: SCEN-plugins-first-party-manifest-label
type: scenario
---
Given plugins/nbp-fx/manifest.json, when it gains an egress host but the Dockerfile's label is not rewritten, then task check fails naming the Dockerfile; after task plugins:labels the label holds the new manifest and the check passes; and a demo Plugin of that image that still declares the old manifest fails the check, since the operator would refuse it.
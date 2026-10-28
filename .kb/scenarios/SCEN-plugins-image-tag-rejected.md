---
title: A plugin image named by tag is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:11.416Z'
id: SCEN-plugins-image-tag-rejected
type: scenario
---
Given the image references registry.example.com/sample:1.0.0, registry.example.com/sample:1.0.0@sha256:<64 hex> and registry.example.com/sample@sha256:<64 hex>, when each is parsed, then only the last is accepted, and its String is the same repository@digest text a Pod spec would use. Assumes cosign verification is a separate adapter step.
---
title: The declared manifest must be the signed label
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:50:54.108Z'
id: SCEN-plugins-manifest-label-match
type: scenario
---
Given an image whose signed manifest label requests the view capability, when a Plugin declares the same manifest in another order then it is accepted, and when the Plugin declares one more egress host or capability, or the label has an unknown field, then the image is refused.
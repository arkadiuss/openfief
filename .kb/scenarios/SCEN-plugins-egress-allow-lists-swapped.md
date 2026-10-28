---
title: A changed allow-list applies whole to the next CONNECT
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:46:28.598Z'
id: SCEN-plugins-egress-allow-lists-swapped
type: scenario
---
Given plugin sample allowed example.com, when its allow-list is replaced by one without example.com, then the next CONNECT to example.com is refused, and when the original list is restored it succeeds again; a ConfigMap update replaces every plugin's list at once.
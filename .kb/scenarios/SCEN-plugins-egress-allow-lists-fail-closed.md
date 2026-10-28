---
title: A broken allow-list ConfigMap denies all egress
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:07.403Z'
id: SCEN-plugins-egress-allow-lists-fail-closed
type: scenario
---
Given the proxy watching openfief-egress-allowlist with a valid list for sample, when the ConfigMap is updated to a value that does not decode, then no plugin has a list and the error is reported; when it is fixed, the list returns; when it is deleted, no plugin has a list again.
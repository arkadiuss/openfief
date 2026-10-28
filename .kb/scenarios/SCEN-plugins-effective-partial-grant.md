---
title: A partial grant enforces only the overlap
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:05.200Z'
id: SCEN-plugins-effective-partial-grant
type: scenario
---
Given a connector requesting api.example.com:443, data.example.org:8443, the Secrets sample-token and sample-cert and data tier T1, and a grant approving api.example.com:443, data.example.org:443 and only sample-token, when effective permissions are computed, then they hold the connector capability, only api.example.com:443 and only sample-token at its requested mount path, and no data tier; and a grant approving only the view capability yields no effective permissions at all.
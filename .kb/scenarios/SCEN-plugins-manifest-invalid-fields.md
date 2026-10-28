---
title: Every invalid manifest field is reported with its name
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:13.549Z'
id: SCEN-plugins-manifest-invalid-fields
type: scenario
---
Given a connector manifest with a wildcard egress host, an IP-literal egress host and a Secret mounted under /var/run/secrets, when it is validated, then validation fails with one field error for egress[1], one for egress[2] and one for the Secret's mountPath, each wrapping its sentinel error, and a manifest with no capability fails with a capabilities error.
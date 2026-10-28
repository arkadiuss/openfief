---
title: The owner sees the trust and every requested permission before confirming
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T23:36:01.840Z'
id: SCEN-plugins-cli-permission-summary
type: scenario
---
Given a signed image whose manifest requests the connector capability, api.example.com:443 and data.example.org:8443, the Secret sample-credentials at /etc/openfief/sample, data tier T1 and a config schema, when the owner runs install, then the CLI shows the instance name, the image, whether it is a verified plugin and which signer earned that, each capability, egress host, Secret with its mount path and data tier, each config property with its type and whether it is required, and the config given.
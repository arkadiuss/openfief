---
title: The EODHD plugin asks only for eodhd.com and its key, and checks its config
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:16.532Z'
id: SCEN-plugins-eodhd-config
type: scenario
---
Given the plugin's manifest.json, when it is parsed as the image label, then it asks for the connector capability, egress to eodhd.com only and the eodhd-credentials Secret holding the owner's API key; given a config with an ISIN with a wrong check digit, an exchange outside Warsaw, LSE, Xetra and Euronext, or an ISIN or index named twice, then the plugin refuses to start with an invalid config error.
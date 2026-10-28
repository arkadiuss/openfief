---
title: An install grants exactly what the signed manifest requests
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:23:50.137Z'
id: SCEN-plugins-cli-install-grants-request
type: scenario
---
Given a signed image whose manifest label requests the connector capability, two egress hosts, two Secrets and a data tier, when it is installed as sample-main with a config, then the Plugin carries that image, that manifest and that config, the PluginGrant approves exactly the requested capability, hosts, Secret names and tier, and nothing is pending approval; an instance name of 57 characters or in upper case is refused before anything is applied.
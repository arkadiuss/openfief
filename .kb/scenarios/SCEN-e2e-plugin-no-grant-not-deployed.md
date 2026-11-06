---
title: A Plugin without a PluginGrant is verified but not deployed until its grant is applied
status: active
tags:
  - acceptance:p10-11
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:36.265Z'
id: SCEN-e2e-plugin-no-grant-not-deployed
type: scenario
---
Given the suite-signed nbp-fx image in the suite registry and an operator that trusts the suite key, when an owner applies its Plugin with no PluginGrant, as plain kubectl would, then the operator verifies the signature but reports Ready False with reason NoGrant and everything the manifest requests as pending approval, and creates no Deployment, pod or allow-list entry and records no deployed version; once the grant is applied, the plugin becomes Ready. Runs on both kind and k3d.
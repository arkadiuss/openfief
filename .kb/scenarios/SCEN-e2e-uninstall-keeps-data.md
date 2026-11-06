---
title: Uninstalling a connector removes its grant and Secrets and keeps its transactions, marked as from a removed instance
status: active
tags:
  - acceptance:p10-18
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:42.790Z'
id: SCEN-e2e-uninstall-keeps-data
type: scenario
---
Given the synthetic connector instance e2e-books that booked a deposit into the demo XTB account, when the owner runs openfief plugin uninstall e2e-books, then its Plugin, PluginGrant, credentials Secret, Deployment, Service, NetworkPolicy, ServiceAccount and allow-list entry are gone, and the read API still lists the deposit from e2e-books, now with connector_instance_removed set. Runs on both kind and k3d.
---
title: Uninstalling a connector instance leaves no pod and no Secret labelled as its own, and keeps every deposit it booked
status: active
tags:
  - acceptance:p10-18
origin:
  kind: agent
  recorded_at: '2026-11-12T19:55:13.264Z'
id: SCEN-e2e-uninstall-leaves-no-pod-or-secret
type: scenario
---
Given the synthetic connector instance e2e-books that booked two deposits on two dates, the second after its Plugin config changed, when the owner runs openfief plugin uninstall e2e-books, then, besides its Plugin, PluginGrant, Deployment and allow-list entry, no pod labelled as the instance's and no Secret labelled openfief.io/instance=e2e-books of any role remains, and the read API lists both deposits with connector_instance_removed set. The data path's instances are uninstalled with the CLI too and checked for the same afterwards. Runs on both kind and k3d.
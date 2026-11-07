---
title: An optional Secret lets the plugin run before the owner creates it
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:00.356Z'
id: SCEN-plugins-optional-secret-mount
type: scenario
---
Given a manifest that asks for a required credentials Secret and an optional ca-bundle Secret, both approved by the grant, when the operator builds the plugin workload, then the ca-bundle Secret is an optional volume and the credentials Secret a required one; when the CLI shows the install and writes and reads the Plugin back, the ca-bundle Secret stays optional and is marked as one the plugin runs without.
---
title: Uninstall keeps a Secret another instance still uses
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:23:52.021Z'
id: SCEN-plugins-cli-uninstall-shared-secrets
type: scenario
---
Given an instance whose manifest and grant name the Secrets a-cert and shared, and another installed instance of the same image that names shared, when the first is uninstalled, then a-cert is deleted and shared is kept and reported, because Secret names come from the image's manifest and instances of one image share them.
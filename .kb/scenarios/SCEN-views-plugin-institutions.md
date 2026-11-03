---
title: The installed plugins name the institutions each connector books for
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T22:19:45.784Z'
id: SCEN-views-plugin-institutions
type: scenario
---
Given a connector whose Plugin resource's manifest declares the institution eodhd, when any read API caller lists the installed plugins, then the plugin carries institutions eodhd, read from spec.manifest.institutions, and the built-in UI's plugins page says that it books for eodhd only.
---
title: A manifest whose text could steer a terminal is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T09:18:08.352Z'
id: SCEN-plugins-manifest-control-characters-refused
type: scenario
---
Given a signed plugin image, when any string of its manifest, including a config schema's property names, titles, descriptions and enum values, holds a C0 or C1 control character, a bidirectional embedding, override or isolate, or invalid UTF-8, then validation refuses the manifest with ErrUnsafeText, so the CLI never offers it and the operator never runs it.
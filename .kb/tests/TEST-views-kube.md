---
title: Installed plugins are read from the operator's Plugin resources through a watched cache, and the removal of a Plugin is told to a handler
status: active
sourceFile: internal/views/kube/plugins_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:26.196Z'
id: TEST-views-kube
type: test
verification_scope: unit
---
Runs Plugins over the client-go fake dynamic client. RecordOf reads the fields the views show of a Plugin the operator wrote, and a missing or mistyped field reads as empty; a Plugin being deleted is listed as such; the cache lists only the namespace's Plugins, picks up a Plugin added after Start, and Start gives up when its context ends. OnRemoved calls its handler with the instance of every Plugin the watch sees deleted, and with none that is only added: the core ends a view plugin's grants when it is uninstalled.
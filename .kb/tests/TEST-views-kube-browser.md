---
title: The core reads which view plugins serve browsers from the operator's Plugins and Services
status: active
sourceFile: internal/views/kube/browser_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:43.954Z'
id: TEST-views-kube-browser
type: test
---
Uses Plugins converted from the operator's own v1alpha1 types and the client-go fakes: only a ready, not deleted plugin with view effective and an effective browser path yields a route; routes come in instance order after Start, ErrViewsUnknown before it; BaseURL is the plugin Service's ClusterIP and connect port, and a headless Service has no address.
---
title: Only running connector instances with a callable Service are synced
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:38.341Z'
id: SCEN-ingest-discover-instances
type: scenario
---
Given Plugin resources and their Services in the openfief namespace, when the core lists instances to sync, then it lists the Plugins that are Ready, not being deleted and allowed the connector capability and whose Service has a ClusterIP and a connect port, and calls them at that address; when the operator creates or removes a Service or a Plugin, then the watch updates the list.
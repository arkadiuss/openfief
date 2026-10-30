---
title: The operator runs nothing while the in-cluster service can't be read, and tries again
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:18.511Z'
id: SCEN-operator-in-cluster-service-reads-service
type: scenario
---
Given a running version whose grant approves an in-cluster service, when the operator can't read a usable Service there (missing, headless, no selector, no such port, a named targetPort), then it tears the plugin down, reports Ready False with ClusterUnavailable and reconciles again within a minute; when it can, the plugin gets the rule.
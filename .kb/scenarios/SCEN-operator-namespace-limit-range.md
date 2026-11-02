---
title: The namespace caps containers that set no resources
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T20:55:34.427Z'
id: SCEN-operator-namespace-limit-range
type: scenario
---
Given the install's cluster-admin base deploy/crds, when it is applied, then the openfief namespace gets the LimitRange openfief-defaults, which gives every container that sets no resources of its own a request of 50m CPU and 64Mi of memory and a memory limit of 1Gi, with no default CPU limit, so a pod applied into the namespace some other way, or the database's pods, cannot grow without a bound.
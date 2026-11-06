---
title: The default install creates no route, no LoadBalancer or NodePort Service and no tunnel
status: active
tags:
  - acceptance:p10-13
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:40.051Z'
id: SCEN-e2e-default-install-exposes-nothing
type: scenario
---
Given OpenFief installed from the release bundle, with the built-in UI installed through the CLI, when the bundle files and the openfief namespace are read, then neither holds an Ingress, a Gateway API Gateway or route, a DaemonSet or a StatefulSet, every Service is ClusterIP without external IPs, no pod uses the host network or host ports, and every pod belongs to the bundle Deployments, a Deployment a Plugin owns, or the openfief-db cluster. Runs on both kind and k3d.
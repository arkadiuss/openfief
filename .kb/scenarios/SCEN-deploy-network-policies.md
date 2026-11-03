---
title: Only their callers reach the database, the egress proxy and the operator
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:32.314Z'
id: SCEN-deploy-network-policies
type: scenario
---
Given the install's bases, when they are built, then the database's NetworkPolicy admits only the core, its own instances and CloudNativePG's operator, so a plugin pod can't open a connection to it; the egress proxy admits only plugin pods and its own egress excludes the private, shared, link-local and loopback ranges, where cluster pods, Services and nodes live; and the operator admits no ingress at all.
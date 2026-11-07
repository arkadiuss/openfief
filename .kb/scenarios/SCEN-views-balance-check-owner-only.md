---
title: Only the owner gets the balance report, through the core's pod
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:11.241Z'
id: SCEN-views-balance-check-owner-only
type: scenario
---
Given an install with balance assertions, when anyone but the owner, such as a script token, a view plugin or a demo session, asks the views for the balance report, then it is refused; when the owner runs openfief check balances, then the CLI runs the core's check balances command in a ready core pod through Kubernetes exec with the owner's kubeconfig, fails before running anything while no core pod is ready, and the report covers only the accounts the authorizer lets the owner view.
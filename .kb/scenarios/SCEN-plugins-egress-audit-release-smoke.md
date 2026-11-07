---
title: The release smoke test audits the egress proxy's log after installing a published release
status: active
tags:
  - acceptance:p10-16
origin:
  kind: agent
  recorded_at: '2026-11-07T03:30:49.982Z'
id: SCEN-plugins-egress-audit-release-smoke
type: scenario
---
Given a published release installed on kind as the install guide says, with the UI plugin and the first passkey, when the smoke test's last step runs egress-proxy audit built from the checkout with the cluster's kubeconfig, then a failed audit fails the smoke test, and the audit runs after every step of the guide.
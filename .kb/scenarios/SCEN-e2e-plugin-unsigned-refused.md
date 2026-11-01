---
title: The operator refuses an unsigned plugin image and runs nothing
status: active
tags:
  - acceptance:p10-11
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:54.089Z'
id: SCEN-e2e-plugin-unsigned-refused
type: scenario
---
Given the operator on a local k3d or kind cluster and an image in a registry it reads that nobody signed, when the owner applies a Plugin for it with a PluginGrant that approves everything its manifest requests, then SignatureVerified is False with reason Unsigned and Ready is False with SignatureNotVerified, no Deployment, pod or allow-list entry exists for it, its status records nothing as deployed, and a Warning event on the Plugin names the refusal. This verifies part of P-10's acceptance criterion 11.
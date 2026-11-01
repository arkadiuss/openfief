---
title: The operator refuses a plugin image signed by a key it does not trust
status: active
tags:
  - acceptance:p10-11
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:58.826Z'
id: SCEN-e2e-plugin-untrusted-signer-refused
type: scenario
---
Given the operator on a local k3d or kind cluster trusting only the suite's key, and an image signed with a second key whose signature verifies over the image digest and its manifest label, when the owner applies a Plugin for it with a PluginGrant that approves everything its manifest requests, then the operator counts no valid signature from a trusted signer: SignatureVerified is False with reason Unsigned, Ready is False with SignatureNotVerified, nothing is deployed and a Warning event names the refusal. This verifies part of P-10's acceptance criterion 11.
---
title: The CLI refuses an image the operator would refuse, with the operator's reason, and applies nothing
status: active
tags:
  - acceptance:p10-11
origin:
  kind: agent
  recorded_at: '2026-11-12T19:55:13.143Z'
id: SCEN-e2e-cli-refusal-matches-operator
type: scenario
---
Given an image nobody signed, one signed by a key the owner does not trust, one named by tag alone, and one naming a tag next to its digest, when an owner runs openfief plugin install on each, then the CLI exits non-zero, no Plugin and no PluginGrant exists, and its output gives the refusal (image has no valid signature, not signed by a trusted signer naming the key, not pinned by digest, invalid image reference). When the same image is then applied with plain kubectl, which the CLI never touches, the operator itself records the same refusal: SignatureVerified False with reason Unsigned, UntrustedSigner or InvalidImage and a message with the CLI's refusal, Ready False, no deployed version, no Deployment and no pod, and for a tag alone the API server refuses the Plugin as invalid so none exists. Runs on both kind and k3d; it assumes the CLI and the operator share the cluster's trusted signers, as D-56 has the CLI preview the operator's check.
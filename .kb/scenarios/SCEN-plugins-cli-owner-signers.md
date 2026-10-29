---
title: The CLI trusts the owner's signers the operator trusts
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T23:14:00.143Z'
id: SCEN-plugins-cli-owner-signers
type: scenario
---
Given the openfief-trusted-signers ConfigMap with a signers.json naming the signer my-ci, when the CLI builds its trust policy, then it trusts my-ci next to OpenFief's CI and the catalog's publishers; with no ConfigMap or no key it adds no signer, and an invalid document stops the install, as it stops the operator.
---
title: The operator reports Verified only for the image the catalog pins, Unverified for an owner-signed one
status: active
tags:
  - acceptance:p10-11
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:38.510Z'
id: SCEN-e2e-plugin-verified-when-catalog-pinned
type: scenario
---
Given the nbp-fx image the embedded catalog pins, copied with its CI signatures from GHCR into the suite registry at the same digest, and the nbp-fx image the checkout built, signed by the owner key, when each is applied as a Plugin without a grant, then the operator verifies both offline and its status calls the catalog image Verified, signed by the catalog publisher, and the owner image Unverified, signed by the owner signer. Assumes the runner reaches ghcr.io. Runs on both kind and k3d.
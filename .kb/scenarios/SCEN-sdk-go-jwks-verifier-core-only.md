---
title: Only the core's token for this instance passes the SDK verifier
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:25.733Z'
id: SCEN-sdk-go-jwks-verifier-core-only
type: scenario
---
Given a plugin instance nbp whose mounted key directory holds the cluster's issuer and public keys, when calls arrive with the core's token for openfief-plugin:nbp, with the core's token for another instance, with another ServiceAccount's token, with an expired token, with an alg none token and with an HMAC token keyed with the public key, then only the first passes and every other is refused as not the core. Assumes the operator keeps the mount current.
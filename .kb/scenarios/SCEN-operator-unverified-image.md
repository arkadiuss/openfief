---
title: An image whose signature is not verified is never deployed
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:59:56.372Z'
id: SCEN-operator-unverified-image
type: scenario
---
Given a granted Plugin and the operator's UnimplementedVerifier, when the operator reconciles it, then no Deployment exists, SignatureVerified is Unknown with reason NotImplemented and Ready is False with reason SignatureNotVerified. Assumes cosign verification will replace the verifier.
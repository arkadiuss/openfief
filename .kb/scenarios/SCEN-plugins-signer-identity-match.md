---
title: A signature counts only for the exact signer identity
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:50:38.019Z'
id: SCEN-plugins-signer-identity-match
type: scenario
---
Given OpenFief's CI identity (the GitHub Actions issuer and the images.yml workflow at refs/heads/main), when a signature's certificate names the same workflow in a fork, on another branch, or with extra text around the subject, then it does not match; and given an owner's public key, a signature by any other key does not match. Assumes the certificate was already verified cryptographically.
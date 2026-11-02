---
title: Release and image verification pins the signing run's repository and ref
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T20:00:09.252Z'
id: SCEN-tooling-verify-pins-source-repository
type: scenario
---
Given the install guide, release-verify.sh, verified-digest.sh, pin-images.sh and release-catalog.sh, when any of them verifies OpenFief's CLI files or images with cosign, then besides the certificate identity and GitHub's issuer it passes --certificate-github-workflow-repository with the repository and --certificate-github-workflow-ref with the ref of the signing run, so a CLI or image signed by OpenFief's workflow identity in another repository's or ref's run fails, and the catalog's OpenFief publisher entries carry the same source repository and ref.
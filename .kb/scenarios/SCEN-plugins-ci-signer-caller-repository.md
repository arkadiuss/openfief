---
title: OpenFief's CI identity counts only from OpenFief's own run on main
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T20:00:04.929Z'
id: SCEN-plugins-ci-signer-caller-repository
type: scenario
---
Given OpenFief's CI as a trusted signer, pinned to the source repository https://github.com/arkadiuss/openfief and the source ref refs/heads/main, when an image carries a valid keyless signature whose certificate has OpenFief's images.yml subject and GitHub's issuer but a source repository of another repository that called the reusable workflow, or OpenFief's repository on another ref, or no source extensions at all, then the signature is not OpenFief CI's, the image is refused as signed by an untrusted signer, and the refusal names the source repository and ref. When the same subject comes from OpenFief's repository on main, the image is accepted as signed by openfief-ci. A broad owner pattern that matches the caller's signature makes it the owner's, never OpenFief CI's.
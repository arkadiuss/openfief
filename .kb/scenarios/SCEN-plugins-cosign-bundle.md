---
title: Only a valid cosign signature over the exact digest counts
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:16:53.891Z'
id: SCEN-plugins-cosign-bundle
type: scenario
---
Given an image in a registry with cosign v3 bundles attached as referrers, when a bundle chains to a trusted Fulcio CA with an SCT, is in a trusted Rekor log with an observer timestamp inside the certificate's validity, signs the image's exact digest with the cosign sign predicate and names a trusted signer, then it counts; a bundle failing any of these, or a config blob that does not match its digest, never counts.
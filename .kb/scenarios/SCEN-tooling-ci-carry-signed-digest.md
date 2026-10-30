---
title: An unchanged image keeps its signed digest
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:03.730Z'
id: SCEN-tooling-ci-carry-signed-digest
type: scenario
---
Given an image whose inputs did not change since the build signed for an earlier commit on main, when CI carries it, then the earlier digest's signature from the images workflow on main is verified first, the digest is tagged sha-<commit> and main unchanged, and the new tag is verified to resolve to the same signed digest, so the demo deploy pins a signed digest for every image; an unsigned digest is never tagged.
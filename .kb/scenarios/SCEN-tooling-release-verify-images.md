---
title: The install's images verify as images.yml on main, on a dry run at main's tag
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:23.386Z'
id: SCEN-tooling-release-verify-images
type: scenario
---
Given the bundle's install file, when the release run verifies it, then each image it runs must be pinned by digest and carry a keyless signature by images.yml on main for that very digest; on a dry run, whose install is pinned to placeholders, each image's main tag is verified with the same identity instead. An unsigned image, a signature for another digest or an image named by tag fails the run.
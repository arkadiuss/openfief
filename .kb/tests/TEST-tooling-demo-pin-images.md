---
title: pin-images.sh writes only digests whose signature verifies and fails while a placeholder is left
status: active
sourceFile: scripts/tests/demo-pin-images.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:06.580Z'
id: TEST-tooling-demo-pin-images
type: test
---
Runs scripts/demo/pin-images.sh on a copy of deploy/demo with a fake cosign that verifies only listed image:tag pairs. It checks that both digests are written, comments survive and no placeholder is left; that cosign is asked for the images workflow identity on main and the GitHub OIDC issuer; that one unsigned image fails the run and keeps its placeholder, so nothing can be published; and that a malformed tag or digest is refused before anything is written. A false pass would be a digest copied without verification: the fake prints nothing for unsigned images.
---
title: The CLI refuses an image the operator would never deploy
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:43:23.318Z'
id: SCEN-plugins-cli-install-refuses-untrusted
type: scenario
---
Given a trust policy holding OpenFief's CI and one catalog publisher, when the owner installs an unsigned image, an image signed only by a stranger, an image whose registry is unreachable or a config the manifest's schema rejects, then nothing is proposed and each fails with its reason; a catalog plugin signed by its publisher is proposed as verified and a CI-signed image the catalog does not pin as unverified, and installing under a name that already exists fails.
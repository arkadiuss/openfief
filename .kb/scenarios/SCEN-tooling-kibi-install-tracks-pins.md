---
title: task kibi:install reinstalls when a kibi pin or a patch is not what is installed
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:41:59.333Z'
id: SCEN-tooling-kibi-install-tracks-pins
type: scenario
---
Given node_modules with kibi packages installed and a stamp from the last install, when an installed kibi package is not at the version package-lock.json pins, or a patch in patches/ was edited, added or removed since the stamp, or the stamp is missing, then the status check fails and says why, so task kibi:install runs npm ci; when every pinned version matches and the patches are the stamped ones, it succeeds and the install is skipped; a stamp is written only for an install that matches the pins; and the post-checkout and post-merge hooks run task kibi:install only when package.json, package-lock.json or patches/ changed between the old and the new HEAD.

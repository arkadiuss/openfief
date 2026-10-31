---
title: The UI installs as an app from its manifest
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T08:07:53.759Z'
id: SCEN-ui-installable
type: scenario
---
Given the owner's browser on the UI's origin over HTTPS or localhost, when a page loads, then the browser finds the web app manifest at /manifest.webmanifest, which the CSP admits, naming OpenFief, standalone display, start URL and scope /, the colors of the design tokens and the Seal's icons at 192 and 512 px with a maskable one, so the browser offers to install it; launched from the home screen the owner signs in with a passkey as in a tab.
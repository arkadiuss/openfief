---
title: A re-run of a failed publish finishes only what is missing
status: active
origin:
  kind: agent
  recorded_at: '2026-11-05T21:54:48.027Z'
id: SCEN-tooling-release-publish-resume
type: scenario
---
Given a publish that already created the tag at the release commit and a release missing some assets, when task release:publish runs again, then it creates no tag and no release, uploads only the assets missing by name, and opens the catalog pull request only when no open or merged one from release/<version> exists; a tag at another commit is refused before anything is written.

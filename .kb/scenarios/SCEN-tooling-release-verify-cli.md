---
title: The release's CLI verifies only as release.yml at the run's ref
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:20.040Z'
id: SCEN-tooling-release-verify-cli
type: scenario
---
Given the CLI files a release run built and signed keylessly, when the run verifies them as the install guide does, then SHA256SUMS must match every file and each file must carry a Sigstore bundle whose certificate names release.yml of the repository at the run's ref, refs/heads/main for a real release, issued by GitHub Actions; a changed file, a missing bundle or a signature by another workflow or ref fails the run.
---
title: release-smoke.sh runs the guide's steps in order, then the egress audit, and fails on each broken one
status: active
sourceFile: scripts/tests/release-smoke.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T16:42:32.755Z'
id: TEST-tooling-release-smoke
type: test
---
Runs scripts/release-smoke.sh against a fake release with stand-ins for gh, cosign, kubectl, node and egress-proxy and a script as the CLI: a good release passes with the steps in the guide's order, then egress-proxy audit, and the link handed to the passkey script; an accepted wrong identity, a tampered SBOM, an unsigned or tagged image, an unpinned CloudNativePG manifest, another CLI version, no link, a failed passkey, an unclean enrol stop or a failed egress audit each fail it. It also checks the guide and the script share their commands and that release.yml calls the workflow.
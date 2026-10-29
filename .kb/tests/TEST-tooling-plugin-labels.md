---
title: Plugin image labels stay equal to their committed manifests
status: active
sourceFile: scripts/tests/plugin-labels.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:50.974Z'
id: TEST-tooling-plugin-labels
type: test
---
The shell test copies plugins/ and deploy/demo into a throwaway root and runs scripts/plugin-labels.sh. The committed labels pass; a manifest changed without its label fails naming the Dockerfile; write brings the label in step and puts the new host in it; a single quote the label can't hold fails; backslashes written survive a check; a Dockerfile without the label line fails write; and an unknown mode prints the usage. A false pass would be a check comparing nothing; the changed-manifest case guards it.
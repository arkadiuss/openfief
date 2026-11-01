---
title: The command line asks before it changes anything and wires install, print, uninstall and list to the use cases
status: active
sourceFile: internal/plugins/cli/cli_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:35:57.840Z'
id: TEST-plugins-cli-run
type: test
verification_scope: unit
---
Table tests in package cli_test run cli.Run with an in-memory cluster, a fake image check (OpenFief's CI signs one synthetic image, an owner's CI another) and an in-memory file map. They check that an install changes the cluster only after y, not after n or no input, and at once with --yes; flags before or after the argument; the owner's signer from the cluster or a --trusted-signers file and its absence; that --print never connects to the cluster and keeps the summary out of the YAML stream; uninstall with and without a shared Secret, declined and unknown; list; and usage errors. A false pass would be a declined install that still creates the instance; the instance list is compared after every case.
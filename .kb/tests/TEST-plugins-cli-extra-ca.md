---
title: openfief plugin install --extra-ca sets the instance setting, and refuses a manifest without the role
status: active
sourceFile: internal/plugins/cli/extra_ca_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:56.303Z'
id: TEST-plugins-cli-extra-ca
type: test
---
TestRunInstallExtraCA runs the command line against a fake cluster and a fake image inspector: with --extra-ca on a connector whose manifest asks for the ca-bundle role, the created Plugin expects a bundle and the summary says the operator reports it missing and restarts the plugin; without the flag the summary says the Secret is not mounted; a view without the role refuses the flag with ErrExtraCANotRequested and installs nothing. The grant approves the bundle Secret either way.
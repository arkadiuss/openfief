---
title: The CLI store writes the CA bundle Secret labelled as the instance and stamps only the Plugin checksum
status: active
sourceFile: internal/plugins/kube/cabundle_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T17:32:41.122Z'
id: TEST-plugins-kube-ca-bundle
type: test
---
Drives kube.Instances against a controller-runtime fake API server. TestInstancesPutCABundle creates the Secret <instance>--ca-bundle with ca.crt and the openfief.io/instance label, replaces its data on a rotation, deletes it twice without error, and refuses with ErrSecretNotOwned to overwrite one labelled as another instance. TestInstancesStampCABundle merge-patches spec.caBundleSha256 and clears it, comparing the whole Plugin read back so nothing else changes, and fails with ErrNotInstalled without a Plugin. TestCABundleYAML pins the --print YAML of the Secret and the partial Plugin. TestInstancesKeepCABundleStamp and TestInstancesKeepExtraCA round-trip the stamp and the setting through the Plugin resource.
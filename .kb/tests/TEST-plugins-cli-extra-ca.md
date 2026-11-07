---
title: The CLI sets extraCa at install or upgrade, and ca-bundle set and unset write the Secret and the stamp
status: active
sourceFile: internal/plugins/cli/extra_ca_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:56.303Z'
id: TEST-plugins-cli-extra-ca
type: test
---
Runs the command line against a fake cluster that records CA bundle writes in order and a fake image inspector. TestRunInstallExtraCA: --extra-ca sets the Plugin setting and the summary says how to set the bundle, without it the summary says not mounted, a manifest without the role refuses the flag. TestRunCABundleSet: set with a self-signed certificate made for the test writes the Secret then the stamp, the SHA-256 of the file; a key, text, a missing file, an instance not expecting a bundle and one not installed are refused with no write; --print needs no cluster, prints the Secret and stamp and the note to keep them in step; bad arguments are usage errors. TestRunCABundleUnset: declining changes nothing, confirming deletes the Secret then clears the stamp. TestRunUpgradeExtraCA: --no-extra-ca and --extra-ca toggle the setting without reinstalling, a plain upgrade keeps setting and stamp, turning it off clears the stamp, and --extra-ca without the role or both flags are refused.
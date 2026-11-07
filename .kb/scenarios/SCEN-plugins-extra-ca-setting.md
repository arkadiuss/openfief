---
title: Only an instance that expects an extra CA bundle mounts it
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:58.862Z'
id: SCEN-plugins-extra-ca-setting
type: scenario
---
Given a connector whose signed manifest asks for the optional ca-bundle Secret role and a grant that approves that Secret, when the owner installs it with --extra-ca then the Plugin carries spec.extraCa and the pod mounts <instance>--ca-bundle; without the flag the Secret is not mounted and the summary says so; a manifest without the role refuses the flag. Later, openfief plugin upgrade with --extra-ca or --no-extra-ca turns it on or off without reinstalling and without a new approval; --no-extra-ca also clears the stamped checksum, and a plain upgrade keeps setting and stamp while the role stays.
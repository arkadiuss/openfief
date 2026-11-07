---
title: Only an instance that expects an extra CA bundle mounts it
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:58.862Z'
id: SCEN-plugins-extra-ca-setting
type: scenario
---
Given a connector whose signed manifest asks for the optional ca-bundle Secret role and a grant that approves that Secret, when the owner installs it with --extra-ca then the Plugin carries spec.extraCa, the pod mounts <instance>--ca-bundle and the install lets the operator get that one Secret; when the owner installs it without the flag the Secret is not mounted and the summary says so; a manifest without the role refuses the flag. Turning the setting on later needs no new approval, and an upgrade keeps it while the new manifest still asks for the role.
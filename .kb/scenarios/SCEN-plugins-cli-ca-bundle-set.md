---
title: ca-bundle set writes the bundle Secret and stamps its checksum
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T17:32:51.822Z'
id: SCEN-plugins-cli-ca-bundle-set
type: scenario
---
Given an installed instance that expects an extra CA bundle, when the owner runs openfief plugin ca-bundle set <instance> <pem-file> with a file of PEM certificates then the CLI creates or replaces the Secret <instance>--ca-bundle with the file under ca.crt, labelled openfief.io/instance=<instance>, and then merge-patches the Plugin spec.caBundleSha256 to the SHA-256 of the file. An empty file, text, a private key or any non-CERTIFICATE block, a certificate that does not parse, an instance that does not expect a bundle, one not installed, and a Secret labelled as another instance are refused before anything changes. With --print it writes the Secret and the partial Plugin stamp as YAML without a cluster, with a note that the owner keeps the checksum in step with the Secret.
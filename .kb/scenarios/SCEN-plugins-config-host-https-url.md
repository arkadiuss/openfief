---
title: Only the host of an https URL in a named config field is requested
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:13.115Z'
id: SCEN-plugins-config-host-https-url
type: scenario
---
Given a manifest whose egressFromConfig names url, when the instance's config sets url to an https URL, then the Plugin requests that URL's host and port (443 when none), lower-cased; when url is absent or empty it requests nothing; when url is plain http, an IP address, a wildcard or has a user part, the Plugin is invalid at config.url.
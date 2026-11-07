---
title: With the HTTPRoute component's auth-origin patch, a passkey created through a one-time link for the route's hostname works there
status: active
tags:
  - acceptance:p10-13
  - acceptance:p10-8
origin:
  kind: agent
  recorded_at: '2026-11-07T03:30:45.452Z'
id: SCEN-e2e-passkey-on-route-hostname
type: scenario
---
Given the bundle install exposed through the HTTPRoute component and a Gateway that terminates TLS for a test hostname with a certificate from a CA made for the run, once every other bundle test has finished, when the component's patch of the core's OPENFIEF_AUTH_ORIGINS is applied with the hostname in place of its placeholder and the core has rolled out, then the route still serves /readyz, `openfief enroll --origin https://<hostname>` prints a one-time link for it without a port-forward, and Chromium, sent to the Gateway for that hostname and trusting only the run's CA and the Gateway's keys, creates a passkey there with its virtual authenticator, keeps the __Host- session cookie, signs out and signs in again with it. Runs on both kind and k3d.
---
title: Signed in on localhost, the settings page's one-time link creates a passkey for a second hostname, which then signs in there
status: active
tags:
  - acceptance:p10-8
origin:
  kind: agent
  recorded_at: '2026-11-08T07:27:28.117Z'
id: SCEN-e2e-passkey-second-host-settings-link
type: scenario
---
Given the bundle install exposed through the HTTPRoute component with its auth-origin patch, so the core accepts http://localhost:8080 and https://<test hostname>, when the owner signs in on localhost with a passkey created from openfief enroll's link, enters the hostname's origin in the settings page's Passkey on another host form and confirms with the passkey if the page asks for a recent check (D-183), then the page shows a one-time link for that origin (D-60), and Chromium, sent to the Gateway for the hostname and trusting only the run's CA and Gateway keys, creates a passkey there with the link, keeps the __Host- session cookie, signs out and signs in again with that passkey, and the settings page lists passkeys on localhost and on the hostname. Runs on both kind and k3d.
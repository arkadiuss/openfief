---
title: Floats in the domain and package state fail the architecture test
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:50.752Z'
id: SCEN-tooling-archtest-float-or-state
type: scenario
---
Given a domain function returning float64, an app struct with a float32 field, a contract field of type double, a domain function with a go statement and a domain regular expression held in a package variable, when the money and pure-domain rules run, then each is reported by package and declaration; an adapter's float, an error sentinel and an app package's goroutine are not.

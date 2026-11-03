---
title: Every page the core proxies from a view carries the core's security headers
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:37.342Z'
id: SCEN-views-proxied-pages-security-headers
type: scenario
---
Given a view plugin behind the core's front door, when it answers with no security headers, with its own stricter content security policy, or with headers that would allow framing, sniffing or referrers, then the browser gets the core's content security policy next to the view's own, X-Frame-Options DENY, X-Content-Type-Options nosniff and Referrer-Policy no-referrer; the core's own refusals carry them too.
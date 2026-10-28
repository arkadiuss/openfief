---
title: The Pages verification TXT record follows GITHUB_PAGES_CHALLENGE
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T15:29:26.562Z'
id: SCEN-tooling-site-dns-challenge
type: scenario
---
Given GITHUB_PAGES_CHALLENGE=0123abcd4567ef, when task dns:site runs, then a TXT record _github-pages-challenge-arkadiuss.example.com with that value and the site's comment is created; a new value replaces it, and the quoted form Cloudflare may return counts as the same value. Given the variable is unset, then the run says the TXT record is skipped and leaves an existing one of ours as it is. A value with characters other than letters, digits, dashes and underscores is refused before any API call.
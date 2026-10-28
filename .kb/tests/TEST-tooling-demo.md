---
title: demo.sh plans, creates and destroys demos by label within the budget, with a fake hcloud
status: active
sourceFile: scripts/tests/demo.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:02.243Z'
id: TEST-tooling-demo
type: test
---
Runs scripts/demo/demo.sh against a fake hcloud that logs every call and answers with synthetic prices and inventories, so no Hetzner API is called. It checks that plan prints each price and the gross total and creates nothing; that up refuses a bigger server or a project whose existing labelled resources push the total over DEMO_BUDGET_EUR, before any create call; that up creates the firewall, then the server with the app and env labels, the labelled SSH keys and the rendered cloud-init, prints the sslip.io URL, and creates nothing when both exist; that the firewall allows SSH only from DEMO_SSH_CIDRS and HTTP and HTTPS from DEMO_HTTP_CIDRS (everyone by default, none closes them); that down needs CONFIRM equal to the environment name and deletes the server before its firewall; that a missing token, SSH key, bad CIDR, bad DEMO_ACME or bad ENV is refused; that the token never appears in arguments, output or files; and that the rendered cloud-init pins versions and checksums per architecture, keeps Traefik and servicelb, writes the sslip.io host and the Let's Encrypt production or staging server, and rejects malformed or missing values. A false pass would be a fake that accepts any call: it fails on unexpected arguments.
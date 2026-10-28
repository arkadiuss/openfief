# Research notes

External facts the design and proposals draw on, as found on the date shown. This is evidence, not design: products, APIs and rules change, so re-verify before relying on any of it. Gathered with Claude (AI assistant), mostly through web search.

## Products and prior art

- **[R-1]** (2026-10-25) No existing product combines portfolio tracking with budget-leased, policy-gated agent execution and per-role model routing. The closest open-source pieces:
  - Wealthfolio: AGPL-3.0, Rust/Tauri, local-first, AI assistant and an MCP server. https://github.com/wealthfolio/wealthfolio
  - Ghostfolio: AGPL-3.0, Angular/NestJS/Postgres, multi-user, sells a hosted Premium tier. https://github.com/ghostfolio/ghostfolio
  - TradingAgents: Apache-2.0, LangGraph multi-agent analysis, no execution, supports Ollama and vLLM. https://github.com/TauricResearch/TradingAgents
  - virattt/ai-hedge-fund: MIT, paper trading by default, principle "the LLM never touches the trade". https://github.com/virattt/ai-hedge-fund/blob/main/VISION.md
- **[R-10]** (2026-10-27) Wealthfolio has a self-hosted web edition: one Rust service serves the React UI and a REST API, with embedded SQLite on a volume (image `wealthfolio/wealthfolio`, port 8088). CSV import is documented only as a UI feature; no documented, stable HTTP API for writing activities was found. Community MCP servers write straight into its SQLite database. Secondary sources only; wealthfolio.app was unreachable.
  - https://railway.com/deploy/wealthfolio-tracker
  - https://dev.co/devops/open-source/wealthfolio
  - https://glama.ai/mcp/servers/@datamoc/WFmcp/blob/811e011a3a13a81f175a8907da99d1fe00cf6b2c/docs/WRITE_OPERATIONS.md
- **[R-11]** (2026-10-27) Ghostfolio: a user's security token is exchanged for a bearer token via `POST /api/v1/auth/anonymous`, and activities are imported via `POST /api/v1/import` with an `activities` array. A community guide says the bearer token lasts about 6 months. Its `MANUAL` data source prices custom assets with a scraper that reads HTML or, via JSONPath, JSON. Prior art: `agusalex/ghostfolio-sync` syncs IBKR Flex Queries into Ghostfolio, maps symbols in a `mapping.yaml`, and warns that data outside the Flex date window is lost if never synced. Secondary sources only; ghostfol.io was unreachable.
  - https://github.com/ghostfolio/ghostfolio
  - https://gethomepage.dev/widgets/services/ghostfolio/
  - https://deepwiki.com/ghostfolio/ghostfolio/5.2-market-data-sources
  - https://hub.docker.com/r/agusalex/ghostfolio-sync
- **[R-12]** (2026-10-27) OpenClaw plugins declare an `openclaw.plugin.json` manifest whose JSON Schema config is validated without running plugin code; the plugin itself is TypeScript loaded in-process and its docs say to treat plugins as trusted code. Its open skill registry, ClawHub, was hit by malware: Koi Security flagged 341 of 2,857 skills as malicious in late January 2026, and later counts ranged from about 900 to over 1,100.
  - https://docs.openclaw.ai/plugins
  - https://thehackernews.com/2026/02/researchers-find-341-malicious-clawhub.html
  - https://www.bitdoze.com/openclaw-security-guide/

## Brokers and data sources

- **[R-2]** (2026-10-25) XTB disabled its public xAPI on 2025-03-14. The only options left are statement import or reverse-engineered clients that breach its terms. http://developers.xstore.pro/
- **[R-3]** (2026-10-25) Interactive Brokers:
  - Access paths: TWS API / IB Gateway, the Client Portal Web API, and Flex Queries for reporting.
  - Its official MCP server (July 2026) only drafts orders; the client reviews and submits them on IBKR's platform. https://www.interactivebrokers.com/en/general/about/mediaRelations/7-28-26.php
  - Linked accounts keep separate cash, positions and buying power.
  - TWS precautionary settings can be bypassed for API orders, so they are not a control. https://www.interactivebrokers.com/docs/third-party-integrations/general-third-party-frequently-asked-questions
- **[R-4]** (2026-10-25) Trading 212's public API has scoped keys (such as `orders:execute`), an optional IP allow-list, and live orders in the main account currency. https://helpcentre.trading212.com/hc/en-us/articles/14584770928157-Trading-212-API-key
- **[R-5]** (2026-10-25) Polish treasury bonds have no API from obligacjeskarbowe.pl or PKO BP. Valuation is deterministic from series terms plus GUS CPI data. Open-source valuers: https://github.com/maciekmm/obligacje (MIT) and https://github.com/krbob/edo-calculator

## Security

- **[R-6]** (2026-10-25) Prompt injection against agents that move money is documented in real incidents, such as the May 2026 Grok/Bankrbot wallet drain. There is research on adversarial news attacks against multi-agent trading systems: https://arxiv.org/pdf/2608.24069
- **[R-13]** (2026-10-27) Hostname-based egress control for pods. Standard Kubernetes NetworkPolicy matches IP blocks, pods and namespaces, never hostnames. The upstream standard for domain names (`AdminNetworkPolicy` `domainNames`, NPEP-133) is still provisional. Domain policies exist in Cilium (`toFQDNs`), in GKE (`FQDNNetworkPolicy`, which cannot tell apart domains served by the same IP) and in Calico's commercial editions. CNI-independent options: `nais/fqdn-policy`, which writes NetworkPolicy `ipBlock`s from DNS, and Stripe's Smokescreen, a Go CONNECT proxy with per-client allow-lists and a replaceable client-identity function.
  - https://network-policy-api.sigs.k8s.io/npeps/npep-133/
  - https://docs.cloud.google.com/kubernetes-engine/docs/how-to/fqdn-network-policies
  - https://docs.tigera.io/calico-cloud/network-policy/domain-based-policy
  - https://github.com/nais/fqdn-policy
  - https://github.com/stripe/smokescreen

## Regulation and tax (not legal advice)

- **[R-7]** (2026-10-25) Running agents that trade other people's money is portfolio management, and giving them personalised buy or sell recommendations is investment advice. Both need a KNF permit under art. 69 of the Polish Act on Trading in Financial Instruments. ESMA treats automated client-facing advice as advice. Running your own software on your own account is not an investment service.
  - https://arslege.pl/dzialalnosc-maklerska/k384/a32931/
  - https://www.esma.europa.eu/sites/default/files/2023-07/ESMA35-43-3861_Supervisory_briefing_on_understanding_the_definition_of_advice_under_MiFID_II.pdf
- **[R-8]** (2026-10-25) ESMA's 2024 statement on AI in investment services expects records of decision processes, data sources and algorithms, and disclosure of chatbot use. https://www.esma.europa.eu/sites/default/files/2024-05/ESMA35-335435667-5924__Public_Statement_on_AI_and_investment_services.pdf
- **[R-9]** (2026-10-25) Foreign brokers (IBKR, Trading 212, Revolut) do not issue PIT-8C, so investors compute PIT-38 themselves: 19% Belka tax, FIFO, and the NBP table A exchange rate from the day before. They also file PIT/ZG per country. Open-source references: https://github.com/volodymyr-kovtun/Pitly and https://klopit.co.pl/

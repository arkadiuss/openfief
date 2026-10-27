# Ubiquitous language

One term, one meaning, one code name. Use these names in code, tests, kibi requirements, commits and docs. Terms come from DESIGN.md and SPEC.md; Go names and banned synonyms are proposals (P) until the author confirms them.

## M0 terms

| Term | Context | Go name (P) | Meaning | Don't say |
|---|---|---|---|---|
| Institution | ingest | `ingest.Institution` | A broker, bank or registry that holds assets: IBKR, XTB, the treasury bond registry. | broker (for the registry) |
| Account | ledger | `ledger.Account` | One account at an Institution, e.g. one IBKR linked account. Belongs to one Sleeve. | portfolio |
| Ledger account | ledger | `ledger.LedgerAccount` | A double-entry account inside an Account: cash per currency, a holding per instrument, fees, withholding tax, FX. | "account" alone |
| Transaction | ledger | `ledger.Transaction` | One economic event from a source. Append-only and versioned; keyed by source plus the institution's own ID (D-18). | trade, activity, operation |
| Posting | ledger | `ledger.Posting` | One leg of a Transaction. A Transaction's postings sum to zero per currency (D-19). | entry, leg |
| Instrument | ledger | `ledger.Instrument` | Something that can be held: identified by ISIN, plus a venue symbol. A bond series is an Instrument. | security, ticker, asset |
| Holding | ledger | `ledger.Holding` | The quantity of one Instrument in one Account on a date, derived from Postings. | position (allowed only inside connector adapters, as the broker's word) |
| Lot | ledger | `ledger.Lot` | A quantity acquired on a date at a cost; consumed FIFO. An opening lot is seeded from a broker's current positions (D-23). | |
| Balance assertion | ledger | `ledger.BalanceAssertion` | A broker-reported balance on a date that the derived balance must match (D-18). | snapshot |
| Sleeve | portfolio | `portfolio.Sleeve` | A part of the portfolio with one role: defensive or offensive (GR-6). | bucket |
| Import batch | ingest | `ingest.Batch` | One run of a connector's ingest path. | sync |
| Price | valuation | `valuation.Price` | An end-of-day price of an Instrument from a market-data source (D-21, D-22). | quote |
| FX rate | valuation | `valuation.FXRate` | A rate to PLN; NBP table A in M0 (D-20). | |
| Valuation | valuation | `valuation.Valuation` | The value of a Holding on a date, with its basis and source. | |
| Valuation basis | valuation | `valuation.Basis` | `market`, `gross_accrued` or `net_if_redeemed` (D-25). | |
| Bond series | valuation | `valuation.Series` | A Polish treasury bond series code, such as EDO0536, with its terms. | bond (ambiguous) |
| Connector | ingest | `ingest.Connector` | A plugin that brings data in through an ingest path (D-7). | integration |
| View | views | `views.View` | A plugin that presents the canonical record elsewhere and never writes back (D-2). | sync, export |
| Read API | views | `views` + `connectapi` | The public, read-only `/v1` API (D-6). | |
| Plugin | plugins | `plugins.Plugin` | The custom resource holding what a plugin requests: image, manifest, config (D-16). | extension, addon |
| Grant | plugins | `plugins.Grant` | The custom resource holding what the owner approved for a plugin (D-16). | permission set |
| Manifest | plugins | `plugins.Manifest` | A plugin's declared capabilities, egress hosts, secrets, data tiers and config schema (D-11). | |
| Egress allow-list | plugins | `plugins.EgressAllowList` | The hostnames a plugin may reach through the egress proxy (D-32). | whitelist |
| Owner | identity | `identity.Owner` | The person who runs an installation (GR-2: one individual investor). | user, customer |
| Script token | identity | `identity.ScriptToken` | A scoped, expiring, read-only personal API token (D-27). | API key |

## Later terms (M1+)

Fief (= lease), Thesis, Agent, Intent, Order, Fill, Policy decision, Audit event, Model portfolio, Target. Meanings are in SPEC.md §3. Don't create code for them before their milestone.

## Rules

- A new domain term goes here first, in the same change as the code that introduces it.
- Broker vocabulary (position, trade, conid, symbol quirks) stays inside connector adapters. Translate at the boundary; that's the anti-corruption layer.
- One name per concept across contexts. If two contexts need different meanings, name them differently; never overload a term.

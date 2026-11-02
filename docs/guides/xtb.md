---
title: Import an XTB account
weight: 40
description: Bring an XTB account in from xStation's "Historia operacji" export, starting from its whole history or from its open positions and cash, map it to an OpenFief account, and confirm the ISIN of each ticker once.
covers:
  - plugins/xtb/**
  - internal/ingest/xtb/**
---

XTB offers no API for a retail account, so OpenFief reads the export xStation writes: the "Historia operacji" (operation history) workbook, one account per file (D-9, R-24). The first-party `xtb` connector maps it, and you upload it in the built-in UI. It needs no egress and no Secret: the file is all it reads.

The export names each instrument by its ticker and market, such as `TESTA.PL`, never by ISIN. OpenFief needs an ISIN for every instrument (D-35), so the first time a ticker appears, the market-data connector looks it up and proposes one, and you confirm it once (D-116). Until you do, that ticker's rows wait, left out of the ledger.

An account starts in one of two ways (D-136): from an export of its whole history, or from the open positions of a current export, which seed one opening lot per position at its open time and price, with the cash balance you enter once. After that, every export you upload adds what came after the start.

```mermaid
flowchart LR
  x["xStation:<br/>Historia operacji .xlsx"] --> u["Upload on the<br/>account's page"]
  u --> s{"how the<br/>account starts"}
  s -- "whole history" --> p["xtb connector:<br/>deposits, withdrawals,<br/>stock rows by ticker"]
  s -- "open positions<br/>and your cash" --> o["xtb connector:<br/>opening lots by ticker,<br/>opening cash"]
  o --> b
  p --> b{"ticker's ISIN<br/>confirmed?"}
  b -- yes --> l["booked"]
  b -- no --> w["waits; the eodhd<br/>connector proposes an ISIN"]
  w --> c["Tickers to confirm:<br/>you confirm or type it"]
  c --> l
```

## Before you start

- An owner session in the built-in UI (passkey sign-in).
- The `xtb` plugin installed, for example `openfief plugin install xtb`, and an account bound to it: create it on the **Accounts** page with the instance as its connector ([Ingest](../concepts/ingest.md#the-owners-accounts)).
- For ISIN proposals, the `eodhd` plugin with your API key ([Valuation](../concepts/valuation.md)). Without it, you type each ISIN yourself.

## 1. Export from xStation

In xStation, open the account's operation history and export it to Excel. For a start from the whole history, choose the whole period since the account was opened; for a start from open positions, today's export is enough, and its `Open Positions` sheet is what counts. Keep the file's name as xStation writes it, `<currency>_<account>_<from>_<to>.xlsx`: the connector reads the account's currency from it, or from the `Open Positions` summary when you rename the file. Never send the file to anyone else; OpenFief stores none of it.

## 2. Start the account

On the account's page, under **Upload statement**, **How the account starts** offers the two starts while the account hasn't started:

- **From the holdings of a current statement (opening).** The connector reads `Open Positions` and makes one opening lot per position: its volume, first acquired on its open date, at a cost of its volume times its open price. The account goes live on the date the report was generated, in UTC, as soon as you upload, even while every position still waits for its ticker (D-160): every later upload is in transactions mode. Enter the account's cash on that date under **Cash at the start**, in the account's currency (the export holds no cash balance), with this upload. You enter the cash once: a later upload with cash is refused. Leave it empty if the account held no cash.
- **From a statement of its whole history (transactions).** Every deposit, buy, sale and dividend from the first deposit on is booked, so cash and holdings add up without any cash entered.

The export gives each open price in the instrument's trading currency without naming it, so an opening lot's cost is booked in the currency of the listing you confirm for its ticker, or, when the listing names none, of the exchange its market suffix names, such as PLN for `.PL` and EUR for `.DE` (D-159). London quotes in pence, pounds and dollars, so for a `.UK` ticker whose listing names no currency you pick it when you confirm the ticker, and its positions wait until you do; none is assumed (D-170). A price in pence (`GBX`) is booked in pounds, divided by 100. A position on a market OpenFief knows no exchange of is listed as a problem.

The first upload names the XTB account it found. If it says the broker account isn't mapped, pick it for this account on the **Accounts** page (D-72) and upload the file again. The XTB account number is kept only for that mapping; the read API never serves it (D-121).

## 3. Confirm the tickers

The upload's result lists the rows left out. Rows that wait for a ticker's ISIN say so, and the result links **Tickers to confirm**. That page lists each ticker with its name at XTB, how many rows wait, and the ISIN the market-data connector proposed, already filled in.

- Check the proposal against the instrument's documents or the exchange's site, then press **Confirm**.
- When no proposal is shown, type the ISIN and press **Save**. A US listing never gets a proposal: the market-data connector covers European exchanges only (D-51).
- A `.UK` ticker also shows a **Trading currency** choice: `GBX` (pence), `GBP` or `USD`. Pick the currency its prices are in at XTB when no proposal names one, or when you type another ISIN than the proposal's; otherwise leave it on the listing's. Without a pick the confirmation is refused, naming the choices, and the rows keep waiting.

Each confirmation books that ticker's waiting rows at once, opening lots too, which join the account's start on its go-live date, and holds for every later upload. It can't be changed afterwards, so check it first.

## 4. Upload again whenever you like

Upload a newer export at any time, without choosing anything: once the account has started, the core uploads in transactions mode. Every row is keyed by XTB's operation ID, so rows already booked stay unchanged (D-18), and only new ones are added. For an account started from open positions, rows dated on or before its go-live are skipped, since the opening lots and cash already hold them (D-23); a trade later on the go-live day itself is skipped too, so start from an export taken after the day's trading. An account started from its whole history is never seeded from open positions afterwards: an opening upload for it is refused.

## What is and isn't imported

| Row | Imported as |
|---|---|
| An open stock or ETF position, in an opening upload | An opening lot of the ticker on the go-live date, once the ticker is confirmed; the cash you enter is the opening cash. Nothing else of the export is read. |
| `Deposit`, `Withdrawal` | A deposit or withdrawal of cash. Its comment, which names your bank or account, is never read. |
| `Stock purchase`, `Stock sell` of a stock or ETF | A buy or sale of the ticker, its units from the comment and its cost or proceeds in the account's currency, once the ticker is confirmed |
| `Dividend` with its `Withholding tax` | A dividend of the ticker with the tax withheld, once the ticker is confirmed |
| Open and closed CFD positions | Flagged as not supported (D-24) |
| `Free funds interest` and its tax, other types | Left out as not supported yet; D-152 books free-funds interest as cash income |

Amounts are rounded to the grosz (or the cent): the export stores them as binary floating-point numbers, and the rounding takes away their last-digit noise. An amount with a real fraction of a grosz is refused as a problem instead. [Ingest](../concepts/ingest.md#mapping-an-xstation-export) explains the mapping in detail.

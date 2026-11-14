package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/identity"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
	attentionv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1/attentionv1connect"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// ownerChange is one procedure that changes something for the owner (D-211):
// how the test drives it through the real core, and the one entry it leaves.
type ownerChange struct {
	// procedure is the Connect procedure, such as the generated constants.
	procedure string
	// action is the audit action of the single entry the call writes.
	action identity.AuditAction
	// drive makes the call as the owner.
	drive func(e *changeEnv) error
	// idempotent says the call, made again, succeeds unless the audit log is
	// down: with the log gone it must answer unavailable.
	idempotent bool
}

// changeEnv is a core with an owner signed in, what the drivers call and
// what they have set up: the demo's bond, an empty bank account and a
// suggested transfer, a flagged balance assertion and the items they raise.
type changeEnv struct {
	t       *testing.T
	srv     *httptest.Server
	session string
	pool    *pgxpool.Pool
	// bondISIN is a bond whose terms are stored.
	bondISIN string
	// suggestedItem and flaggedItem are the numbers of the attention items a
	// suggested transfer and a flagged balance assertion raised.
	suggestedItem, flaggedItem int64
	pair                       *ledgerv1.ConfirmTransferRequest

	// The owner services' clients, which present the built-in UI's token as
	// its calls to the core do.
	settings    valuationv1connect.SettingsServiceClient
	marketData  valuationv1connect.MarketDataServiceClient
	checks      valuationv1connect.PriceCheckServiceClient
	sleeves     portfoliov1connect.SleeveServiceClient
	models      portfoliov1connect.ModelPortfolioServiceClient
	accounts    ingestv1connect.AccountServiceClient
	checkpoints ledgerv1connect.CashCheckpointServiceClient
	transfers   ledgerv1connect.TransferServiceClient
	instruments ledgerv1connect.InstrumentServiceClient
	items       attentionv1connect.AttentionServiceClient
	agents      modelsv1connect.ModelSettingsServiceClient
}

// withClients sets the owner services' clients of e on srv.
func (e *changeEnv) withClients() {
	h, u, bearer := e.srv.Client(), e.srv.URL, connect.WithInterceptors(viewBearer(uiCaller))
	e.settings = valuationv1connect.NewSettingsServiceClient(h, u, bearer)
	e.marketData = valuationv1connect.NewMarketDataServiceClient(h, u, bearer)
	e.checks = valuationv1connect.NewPriceCheckServiceClient(h, u, bearer)
	e.sleeves = portfoliov1connect.NewSleeveServiceClient(h, u, bearer)
	e.models = portfoliov1connect.NewModelPortfolioServiceClient(h, u, bearer)
	e.accounts = ingestv1connect.NewAccountServiceClient(h, u, bearer)
	e.checkpoints = ledgerv1connect.NewCashCheckpointServiceClient(h, u, bearer)
	e.transfers = ledgerv1connect.NewTransferServiceClient(h, u, bearer)
	e.instruments = ledgerv1connect.NewInstrumentServiceClient(h, u, bearer)
	e.items = attentionv1connect.NewAttentionServiceClient(h, u, bearer)
	e.agents = modelsv1connect.NewModelSettingsServiceClient(h, u, bearer)
}

// outcome is the error of a call, wrapped, whatever it answered: the drivers
// make a call and only say whether it failed.
func outcome(_ any, err error) error {
	if err != nil {
		return fmt.Errorf("the call: %w", err)
	}
	return nil
}

// The owner's changes the test drives, in an order that sets up what a later
// one needs: the Bank account before its checkpoint, the Gold kind before its
// sleeve. Each is of the Connect procedure it names; a procedure may appear
// with more than one case.
func ownerChanges() []ownerChange {
	return slices.Concat(settingChanges(), marketDataChanges(), sleeveChanges(), accountChanges(), ledgerChanges(), attentionChanges(), agentChanges())
}

// settingChanges are the owner's settings of the valuation context and the
// model portfolio.
func settingChanges() []ownerChange {
	const origin = localUI
	return []ownerChange{
		{valuationv1connect.SettingsServiceSetNumberFormatProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.settings.SetNumberFormat(e.t.Context(), connect.NewRequest(&valuationv1.SetNumberFormatRequest{Origin: origin, SessionToken: e.session, Key: "comma-dot"})))
		}, true},
		{valuationv1connect.SettingsServiceSetMarginalRateProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.settings.SetMarginalRate(e.t.Context(), connect.NewRequest(&valuationv1.SetMarginalRateRequest{Origin: origin, SessionToken: e.session, Rate: "0.32"})))
		}, true},
		{valuationv1connect.SettingsServiceSetExemptionClassProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.settings.SetExemptionClass(e.t.Context(), connect.NewRequest(&valuationv1.SetExemptionClassRequest{
				Origin: origin, SessionToken: e.session, Isin: "PLTEST000010", ExemptionClass: "investment",
			})))
		}, true},
		{valuationv1connect.SettingsServiceSetEarlyRedemptionFeeProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.settings.SetEarlyRedemptionFee(e.t.Context(), connect.NewRequest(&valuationv1.SetEarlyRedemptionFeeRequest{
				Origin: origin, SessionToken: e.session, Isin: e.bondISIN, Fee: "2.00",
			})))
		}, true},
		{valuationv1connect.SettingsServiceSetChartStartProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.settings.SetChartStart(e.t.Context(), connect.NewRequest(&valuationv1.SetChartStartRequest{Origin: origin, SessionToken: e.session, Start: "2026-01-01"})))
		}, true},
		{valuationv1connect.SettingsServiceSetChartStartProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.settings.SetChartStart(e.t.Context(), connect.NewRequest(&valuationv1.SetChartStartRequest{
				Origin: origin, SessionToken: e.session, AccountId: string(demo.BondsAccount), Start: "2026-02-01",
			})))
		}, true},
		{valuationv1connect.PriceCheckServiceAcknowledgePriceCheckProcedure, identity.ActionPriceCheckAcknowledged, func(e *changeEnv) error {
			return outcome(e.checks.AcknowledgePriceCheck(e.t.Context(), connect.NewRequest(&valuationv1.AcknowledgePriceCheckRequest{
				Origin: origin, SessionToken: e.session, Isin: checkISIN, Symbol: checkSymbol, Date: "2026-03-08", Source: "ibkr-main",
			})))
		}, false},
		{valuationv1connect.PriceCheckServiceSetPriceCheckThresholdProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.checks.SetPriceCheckThreshold(e.t.Context(), connect.NewRequest(&valuationv1.SetPriceCheckThresholdRequest{
				Origin: origin, SessionToken: e.session, Isin: checkISIN, Symbol: checkSymbol, Threshold: "0.05",
			})))
		}, true},
		{portfoliov1connect.ModelPortfolioServiceSaveModelPortfolioProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			model := &portfoliov1.ModelPortfolio{Enabled: true, Targets: []*portfoliov1.Target{
				{Kind: portfoliov1.TargetKind_TARGET_KIND_INSTRUMENT, Isin: "IETEST000026", Symbol: "ETF1", WeightPercent: "70"},
				{Kind: portfoliov1.TargetKind_TARGET_KIND_INDEX, Index: "FAKE20", Currency: "PLN", WeightPercent: "30"},
			}}
			return outcome(e.models.SaveModelPortfolio(e.t.Context(), connect.NewRequest(&portfoliov1.SaveModelPortfolioRequest{
				Origin: origin, SessionToken: e.session, Model: model,
			})))
		}, true},
	}
}

// marketDataChanges are the owner's rankings, series sources and gap
// thresholds.
func marketDataChanges() []ownerChange {
	const origin = localUI
	return []ownerChange{
		{valuationv1connect.MarketDataServiceSetSourceRankingProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.marketData.SetSourceRanking(e.t.Context(), connect.NewRequest(&valuationv1.SetSourceRankingRequest{
				Origin: origin, SessionToken: e.session, Kind: valuationv1.MarketDataKind_MARKET_DATA_KIND_FX_RATES, Sources: []string{"ecb-fx", "nbp-fx"},
			})))
		}, true},
		{valuationv1connect.MarketDataServiceSetSeriesSourceProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.marketData.SetSeriesSource(e.t.Context(), connect.NewRequest(&valuationv1.SetSeriesSourceRequest{
				Origin: origin, SessionToken: e.session, Kind: valuationv1.MarketDataKind_MARKET_DATA_KIND_FX_RATES, Key: "USD", Source: "synthetic",
			})))
		}, true},
		{valuationv1connect.MarketDataServiceSetGapThresholdProcedure, identity.ActionSettingsChanged, func(e *changeEnv) error {
			return outcome(e.marketData.SetGapThreshold(e.t.Context(), connect.NewRequest(&valuationv1.SetGapThresholdRequest{
				Origin: origin, SessionToken: e.session, Kind: valuationv1.MarketDataKind_MARKET_DATA_KIND_PRICES, Days: 14,
			})))
		}, true},
	}
}

// sleeveChanges are the owner's sleeves and sleeve kinds.
func sleeveChanges() []ownerChange {
	const origin = localUI
	return []ownerChange{
		{portfoliov1connect.SleeveServiceAddSleeveKindProcedure, identity.ActionSleeveKindAdded, func(e *changeEnv) error {
			return outcome(e.sleeves.AddSleeveKind(e.t.Context(), connect.NewRequest(&portfoliov1.AddSleeveKindRequest{
				Origin: origin, SessionToken: e.session, Name: "Gold", DefaultRole: portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE,
			})))
		}, false},
		{portfoliov1connect.SleeveServiceRenameSleeveKindProcedure, identity.ActionSleeveKindRenamed, func(e *changeEnv) error {
			return outcome(e.sleeves.RenameSleeveKind(e.t.Context(), connect.NewRequest(&portfoliov1.RenameSleeveKindRequest{
				Origin: origin, SessionToken: e.session, Key: "gold", Name: "Gold and silver",
			})))
		}, false},
		{portfoliov1connect.SleeveServiceAddSleeveProcedure, identity.ActionSleeveAdded, func(e *changeEnv) error {
			return outcome(e.sleeves.AddSleeve(e.t.Context(), connect.NewRequest(&portfoliov1.AddSleeveRequest{
				Origin: origin, SessionToken: e.session, Name: "Gold bars", Kind: "gold",
			})))
		}, false},
		{portfoliov1connect.SleeveServiceUpdateSleeveProcedure, identity.ActionSleeveChanged, func(e *changeEnv) error {
			return outcome(e.sleeves.UpdateSleeve(e.t.Context(), connect.NewRequest(&portfoliov1.UpdateSleeveRequest{
				Origin: origin, SessionToken: e.session, Id: "gold-bars", Name: "Gold coins", Kind: "gold",
			})))
		}, false},
		{portfoliov1connect.SleeveServiceDeleteSleeveProcedure, identity.ActionSleeveDeleted, func(e *changeEnv) error {
			return outcome(e.sleeves.DeleteSleeve(e.t.Context(), connect.NewRequest(&portfoliov1.DeleteSleeveRequest{
				Origin: origin, SessionToken: e.session, Id: "gold-bars",
			})))
		}, false},
		{portfoliov1connect.SleeveServiceDeleteSleeveKindProcedure, identity.ActionSleeveKindDeleted, func(e *changeEnv) error {
			return outcome(e.sleeves.DeleteSleeveKind(e.t.Context(), connect.NewRequest(&portfoliov1.DeleteSleeveKindRequest{
				Origin: origin, SessionToken: e.session, Key: "gold",
			})))
		}, false},
	}
}

// accountChanges are the owner's accounts and their connectors: the Bank
// account, the IBKR main account mapped to a broker account the connector
// reported, and a spare account the owner deletes.
func accountChanges() []ownerChange {
	const origin = localUI
	return []ownerChange{
		{ingestv1connect.AccountServiceCreateAccountProcedure, identity.ActionAccountCreated, func(e *changeEnv) error {
			return outcome(e.accounts.CreateAccount(e.t.Context(), connect.NewRequest(&ingestv1.CreateAccountRequest{
				Origin: origin, SessionToken: e.session, Name: "Bank", Sleeve: "satellite", Institution: "bank",
			})))
		}, false},
		{ingestv1connect.AccountServiceCreateAccountProcedure, identity.ActionAccountCreated, func(e *changeEnv) error {
			return outcome(e.accounts.CreateAccount(e.t.Context(), connect.NewRequest(&ingestv1.CreateAccountRequest{
				Origin: origin, SessionToken: e.session, Name: "IBKR main", Sleeve: "satellite",
				ConnectorInstance: "ibkr-main", BrokerAccount: "U0000001", TaxWrapper: "ikze",
				FeeRule: &ingestv1.FeeRule{Rate: "0.0005", Minimum: "1", MinimumCurrency: "USD"},
			})))
		}, false},
		{ingestv1connect.AccountServiceUpdateAccountProcedure, identity.ActionAccountChanged, func(e *changeEnv) error {
			return outcome(e.accounts.UpdateAccount(e.t.Context(), connect.NewRequest(&ingestv1.UpdateAccountRequest{
				Origin: origin, SessionToken: e.session, AccountId: "bank", Name: "Bank two", Sleeve: "satellite",
			})))
		}, true},
		{ingestv1connect.AccountServiceBindAccountProcedure, identity.ActionAccountBound, func(e *changeEnv) error {
			return outcome(e.accounts.BindAccount(e.t.Context(), connect.NewRequest(&ingestv1.BindAccountRequest{
				Origin: origin, SessionToken: e.session, AccountId: "bank", ConnectorInstance: "ibkr-main", BrokerAccount: "U0000002",
			})))
		}, true},
		{ingestv1connect.AccountServiceRemapBrokerAccountProcedure, identity.ActionBrokerAccountRemapped, func(e *changeEnv) error {
			return outcome(e.accounts.RemapBrokerAccount(e.t.Context(), connect.NewRequest(&ingestv1.RemapBrokerAccountRequest{
				Origin: origin, SessionToken: e.session, ConnectorInstance: "ibkr-main", BrokerAccount: "U0000002",
			})))
		}, true},
		{ingestv1connect.AccountServiceDeleteAccountProcedure, identity.ActionAccountDeleted, func(e *changeEnv) error {
			return outcome(e.accounts.DeleteAccount(e.t.Context(), connect.NewRequest(&ingestv1.DeleteAccountRequest{
				Origin: origin, SessionToken: e.session, AccountId: "spare",
			})))
		}, true},
	}
}

// ledgerChanges are the owner's cash checkpoints, instrument names and
// decisions on transfers.
func ledgerChanges() []ownerChange {
	const origin = localUI
	return []ownerChange{
		{ledgerv1connect.CashCheckpointServiceRecordCashCheckpointProcedure, identity.ActionCashCheckpointRecorded, func(e *changeEnv) error {
			return outcome(e.checkpoints.RecordCashCheckpoint(e.t.Context(), connect.NewRequest(&ledgerv1.RecordCashCheckpointRequest{
				Origin: origin, SessionToken: e.session, AccountId: "bank", Date: "2026-03-31", Currency: "PLN", Amount: "250.00",
			})))
		}, true},
		{ledgerv1connect.CashCheckpointServiceBookCashDifferenceProcedure, identity.ActionCashDifferenceBooked, func(e *changeEnv) error {
			return outcome(e.checkpoints.BookCashDifference(e.t.Context(), connect.NewRequest(&ledgerv1.BookCashDifferenceRequest{
				Origin: origin, SessionToken: e.session, AccountId: "bank", Date: "2026-03-31", Currency: "PLN",
			})))
		}, false},
		{ledgerv1connect.InstrumentServiceRenameInstrumentProcedure, identity.ActionInstrumentRenamed, func(e *changeEnv) error {
			return outcome(e.instruments.RenameInstrument(e.t.Context(), connect.NewRequest(&ledgerv1.RenameInstrumentRequest{
				Origin: origin, SessionToken: e.session, Isin: "PLTEST000010", Name: "Synthetic fund",
			})))
		}, true},
		{ledgerv1connect.TransferServiceConfirmTransferProcedure, identity.ActionTransferConfirmed, func(e *changeEnv) error {
			return outcome(e.transfers.ConfirmTransfer(e.t.Context(), connect.NewRequest(e.pair)))
		}, true},
		{ledgerv1connect.TransferServiceRejectTransferProcedure, identity.ActionTransferRejected, func(e *changeEnv) error {
			return outcome(e.transfers.RejectTransfer(e.t.Context(), connect.NewRequest(&ledgerv1.RejectTransferRequest{
				Origin: origin, SessionToken: e.session, Withdrawal: e.pair.GetWithdrawal(), Deposit: e.pair.GetDeposit(),
			})))
		}, true},
	}
}

// attentionChanges are the owner's decisions on attention items.
func attentionChanges() []ownerChange {
	const origin = localUI
	return []ownerChange{
		{attentionv1connect.AttentionServiceAcknowledgeAttentionItemProcedure, identity.ActionAttentionAcknowledged, func(e *changeEnv) error {
			return outcome(e.items.AcknowledgeAttentionItem(e.t.Context(), connect.NewRequest(&attentionv1.AcknowledgeAttentionItemRequest{
				Origin: origin, SessionToken: e.session, ItemId: e.suggestedItem,
			})))
		}, true},
		{attentionv1connect.AttentionServiceResolveAttentionItemProcedure, identity.ActionAttentionResolved, func(e *changeEnv) error {
			return outcome(e.items.ResolveAttentionItem(e.t.Context(), connect.NewRequest(&attentionv1.ResolveAttentionItemRequest{
				Origin: origin, SessionToken: e.session, ItemId: e.flaggedItem,
			})))
		}, true},
	}
}

// agentChanges are the owner's agents-and-models settings, which the models
// context audited before this change.
func agentChanges() []ownerChange {
	const origin = localUI
	return []ownerChange{
		{modelsv1connect.ModelSettingsServiceSetAgentsEnabledProcedure, identity.ActionModelSettingsChanged, func(e *changeEnv) error {
			return outcome(e.agents.SetAgentsEnabled(e.t.Context(), connect.NewRequest(&modelsv1.SetAgentsEnabledRequest{Origin: origin, SessionToken: e.session, Enabled: true})))
		}, true},
		{modelsv1connect.ModelSettingsServiceSetFeatureEnabledProcedure, identity.ActionModelSettingsChanged, func(e *changeEnv) error {
			return outcome(e.agents.SetFeatureEnabled(e.t.Context(), connect.NewRequest(&modelsv1.SetFeatureEnabledRequest{
				Origin: origin, SessionToken: e.session, Feature: "ask", Enabled: true,
			})))
		}, true},
		{modelsv1connect.ModelSettingsServiceSetRouteProcedure, identity.ActionModelSettingsChanged, func(e *changeEnv) error {
			return outcome(e.agents.SetRoute(e.t.Context(), connect.NewRequest(&modelsv1.SetRouteRequest{Origin: origin, SessionToken: e.session, Role: "ask"})))
		}, true},
		{modelsv1connect.ModelSettingsServiceSetTierModelsProcedure, identity.ActionModelSettingsChanged, func(e *changeEnv) error {
			return outcome(e.agents.SetTierModels(e.t.Context(), connect.NewRequest(&modelsv1.SetTierModelsRequest{Origin: origin, SessionToken: e.session, Tier: "T0"})))
		}, true},
	}
}

// coveredElsewhere are the procedures that change state for the owner (or
// for a caller) and write their entries, whose entries another test pins:
// the test named and the reason. The statement upload, the bond lots and the
// tickers need a connector or a pending ticker the core has only in a cluster
// and with a connector's delivery; the sync retry needs the scheduler, which
// runs only in a cluster; the identity calls and the maintainer's decisions
// are the identity and portfolio contexts' own, with their own tests.
func coveredElsewhere() map[string]struct{ test, why string } {
	const ingestApp = "TestIngestOwnerChangesAreAudited"
	return map[string]struct{ test, why string }{
		ingestv1connect.StatementServiceUploadStatementProcedure: {"TestStatementService", "a flex connector's fake and an opening upload"},
		ingestv1connect.BondLotServiceAddBondLotProcedure:        {"TestBondLotService", "a bonds connector's fake"},
		ingestv1connect.BondLotServiceDeleteBondLotProcedure:     {"TestBondLotService", "a bonds connector's fake"},
		ingestv1connect.TickerServiceConfirmTickerProcedure:      {ingestApp, "tickers waiting for the owner come from a connector's delivery"},
		ingestv1connect.TickerServiceCorrectTickerProcedure:      {ingestApp, "a confirmed ticker comes from a connector's delivery"},
		ingestv1connect.SyncServiceRetrySyncProcedure:            {ingestApp, "the scheduler runs only in a cluster"},
	}
}

// auditedByTheirOwn are the procedures whose entries a context writes by
// its own use case, before this change, each with the test that pins them
// and what it pins: "TestName: entries".
func auditedByTheirOwn() map[string]string {
	const trail = "TestAuditTrail: "
	return map[string]string{
		identityv1connect.AuthServiceFinishRegistrationProcedure:     trail + "passkey_added, enrollment_link_used",
		identityv1connect.AuthServiceFinishLoginProcedure:            trail + "sign_in",
		identityv1connect.AuthServiceCreateDemoSessionProcedure:      trail + "sign_in",
		identityv1connect.AuthServiceLogoutProcedure:                 trail + "sign_out",
		identityv1connect.AuthServiceSignOutEverywhereProcedure:      trail + "sign_out_everywhere",
		identityv1connect.AuthServiceRemovePasskeyProcedure:          trail + "passkey_removed",
		identityv1connect.AuthServiceCreateScriptTokenProcedure:      trail + "script_token_created",
		identityv1connect.AuthServiceRevokeScriptTokenProcedure:      trail + "script_token_revoked",
		identityv1connect.AuthServiceCreateEnrollmentLinkProcedure:   trail + "enrollment_link_issued",
		identityv1connect.AuthServiceInviteViewerProcedure:           trail + "enrollment_link_issued",
		identityv1connect.AuthServiceCancelInviteProcedure:           "TestPendingInvitesAndCancel: enrollment_link_cancelled",
		identityv1connect.AuthServiceRevokePersonProcedure:           "TestRevokeViewer: person_revoked, grant_revoked, passkey_removed",
		identityv1connect.AuthServiceCreateGrantProcedure:            "TestCreateGrant: grant_created",
		identityv1connect.AuthServiceRevokeGrantProcedure:            "TestRevokeGrant: grant_revoked",
		identityv1connect.AuthServiceUpdateGrantsProcedure:           "TestUpdateGrants: grant_created, grant_revoked",
		modelsv1connect.AskModelServiceGenerateAskProcedure:          "TestRouterAuditsBeforeCalling: model_call, model_call_result",
		portfoliov1connect.MaintainerServiceAskMaintainerProcedure:   "TestMaintainerRunsOnARealCore: maintainer_run",
		portfoliov1connect.MaintainerServiceApplyProposalProcedure:   "TestApplyChangesTheTargetsAndTradesNothing: maintainer_proposal_applied",
		portfoliov1connect.MaintainerServiceDismissProposalProcedure: "TestDismissLeavesTheModelAndIsAudited: maintainer_proposal_dismissed",
	}
}

// unauditedByDesign are the procedures that begin or check something and
// change no record of the owner's: a ceremony kept for minutes, a fresh
// check of the owner's passkey. Their finishing calls above are audited.
func unauditedByDesign() map[string]string {
	return map[string]string{
		identityv1connect.AuthServiceBeginRegistrationProcedure:  "stores a ceremony that expires in minutes; FinishRegistration is audited",
		identityv1connect.AuthServiceBeginLoginProcedure:         "stores a ceremony that expires in minutes; FinishLogin is audited",
		identityv1connect.AuthServiceBeginPasskeyCheckProcedure:  "stores a ceremony that expires in minutes; it changes nothing the owner keeps",
		identityv1connect.AuthServiceFinishPasskeyCheckProcedure: "stamps the session as freshly checked, which changes no record the owner keeps; a refused check is no change",
	}
}

// readOnlyProcedures are the procedures that only read, which no entry
// records: the read API is not audited (D-211, author's answer to Q5), and
// reading the owner's own settings and lists changes nothing.
func readOnlyProcedures() []string {
	return []string{
		identityv1connect.AuthServiceGetLoginMethodsProcedure, identityv1connect.AuthServiceWhoAmIProcedure,
		identityv1connect.AuthServiceListPasskeysProcedure, identityv1connect.AuthServiceListScriptTokensProcedure,
		identityv1connect.AuthServiceListAuditEntriesProcedure, identityv1connect.AuthServiceListGrantsProcedure,
		identityv1connect.AuthServiceListViewPluginsProcedure, identityv1connect.AuthServiceListInvitesProcedure,
		identityv1connect.AuthServiceListPeopleProcedure,
		valuationv1connect.SettingsServiceGetOwnerSettingsProcedure,
		valuationv1connect.MarketDataServiceListSourceRankingsProcedure, valuationv1connect.MarketDataServiceListSeriesSourcesProcedure,
		portfoliov1connect.ModelPortfolioServiceGetModelPortfolioProcedure,
		portfoliov1connect.MaintainerServiceGetMaintainerProcedure, portfoliov1connect.MaintainerServiceGetProposalProcedure,
		ingestv1connect.AccountServiceListBrokerAccountsProcedure,
		ingestv1connect.StatementServiceListStatementAccountsProcedure,
		ingestv1connect.TickerServiceListPendingTickersProcedure, ingestv1connect.TickerServiceListConfirmedTickersProcedure,
		modelsv1connect.AskModelServiceGetAskModelProcedure, modelsv1connect.AskModelServiceCheckAskTiersProcedure,
		modelsv1connect.ModelSettingsServiceGetModelSettingsProcedure, modelsv1connect.ModelSettingsServiceListModelProvidersProcedure,
	}
}

// ownerServices are the services of the contracts that the core serves to
// the built-in UI with the owner's session, and whose procedures may change
// something: the owner's, not the read API (no entry for reading), a
// connector's (the core is its caller) nor a model provider's.
func ownerServices() []protoreflect.ServiceDescriptors {
	return []protoreflect.ServiceDescriptors{
		identityv1.File_openfief_identity_v1_auth_proto.Services(),
		valuationv1.File_openfief_valuation_v1_settings_proto.Services(), valuationv1.File_openfief_valuation_v1_market_data_proto.Services(),
		valuationv1.File_openfief_valuation_v1_price_checks_proto.Services(),
		attentionv1.File_openfief_attention_v1_attention_proto.Services(),
		portfoliov1.File_openfief_portfolio_v1_sleeves_proto.Services(), portfoliov1.File_openfief_portfolio_v1_model_proto.Services(),
		portfoliov1.File_openfief_portfolio_v1_maintainer_proto.Services(),
		ledgerv1.File_openfief_ledger_v1_transfers_proto.Services(), ledgerv1.File_openfief_ledger_v1_cash_checkpoints_proto.Services(),
		ledgerv1.File_openfief_ledger_v1_instruments_proto.Services(),
		modelsv1.File_openfief_models_v1_ask_proto.Services(), modelsv1.File_openfief_models_v1_settings_proto.Services(),
		ingestv1.File_openfief_ingest_v1_accounts_proto.Services(), ingestv1.File_openfief_ingest_v1_statements_proto.Services(),
		ingestv1.File_openfief_ingest_v1_tickers_proto.Services(), ingestv1.File_openfief_ingest_v1_syncs_proto.Services(),
		ingestv1.File_openfief_ingest_v1_bond_lots_proto.Services(),
	}
}

// procedures returns the procedure of every method of the owner services.
func procedures() []string {
	var out []string
	for _, services := range ownerServices() {
		for i := range services.Len() {
			svc := services.Get(i)
			for j := range svc.Methods().Len() {
				out = append(out, fmt.Sprintf("/%s/%s", svc.FullName(), svc.Methods().Get(j).Name()))
			}
		}
	}
	return out
}

// listsOf returns how many of the lists of audited, covered, own, unaudited
// and read-only procedures hold p.
func listsOf(p string, driven, readOnly map[string]bool) int {
	n := 0
	for _, in := range []bool{
		driven[p], coveredElsewhere()[p].test != "", auditedByTheirOwn()[p] != "", unauditedByDesign()[p] != "", readOnly[p],
	} {
		if in {
			n++
		}
	}
	return n
}

// TestEveryOwnerProcedureIsClassifiedForTheAuditLog is the guard of D-211: a
// procedure of an owner service that is new must be put in one of the lists
// above, which says whether it is driven here, audited and pinned by another
// test, unaudited by design, or only reads. A procedure that changes
// something and is in none of them fails this test, so no change of state is
// added without an entry or a reason it has none.
func TestEveryOwnerProcedureIsClassifiedForTheAuditLog(t *testing.T) {
	t.Parallel()
	driven, readOnly := map[string]bool{}, map[string]bool{}
	for _, c := range ownerChanges() {
		driven[c.procedure] = true
		if _, err := identity.ParseAuditAction(string(c.action)); err != nil {
			t.Errorf("%s names the action %q, which the identity context doesn't know", c.procedure, c.action)
		}
	}
	for _, p := range readOnlyProcedures() {
		readOnly[p] = true
	}
	all := procedures()
	const atLeast = 80
	if len(all) < atLeast {
		t.Fatalf("found %d owner procedures, want the %d and more of the contracts: the scan broke", len(all), atLeast)
	}
	for _, p := range all {
		if n := listsOf(p, driven, readOnly); n != 1 {
			t.Errorf("%s is in %d of the lists of audited, covered, own, unaudited and read-only procedures, want exactly one: "+
				"a procedure that changes something writes an audit entry (D-211), and one that doesn't says why here", p, n)
		}
	}
	for p := range driven {
		if !slices.Contains(all, p) {
			t.Errorf("%s is driven here but is no procedure of the owner services", p)
		}
	}
}

// declaredTests returns, of names, those a _test.go file of the repository
// declares as a function.
func declaredTests(t *testing.T, names map[string]bool) map[string]bool {
	t.Helper()
	root := os.DirFS("../..")
	skip := map[string]bool{"node_modules": true, ".git": true, ".bin": true, "worktrees": true, "site": true}
	found := map[string]bool{}
	err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return fmt.Errorf("walking %s: %w", path, err)
		case d.IsDir() && skip[d.Name()]:
			return fs.SkipDir
		case d.IsDir() || !strings.HasSuffix(path, "_test.go"):
			return nil
		}
		b, readErr := fs.ReadFile(root, path)
		if readErr != nil {
			return fmt.Errorf("reading %s: %w", path, readErr)
		}
		for name := range names {
			found[name] = found[name] || strings.Contains(string(b), "func "+name+"(")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scanning the tests: %v", err)
	}
	return found
}

// TestCoveredElsewhereNamesTestsThatExist checks that each test the lists
// above rely on is a test function of this repository, so a claim that
// another test pins an entry does not go stale.
func TestCoveredElsewhereNamesTestsThatExist(t *testing.T) {
	t.Parallel()
	names := map[string]bool{}
	for _, c := range coveredElsewhere() {
		names[c.test] = true
	}
	for _, text := range auditedByTheirOwn() {
		name, _, _ := strings.Cut(text, ":")
		names[name] = true
	}
	found := declaredTests(t, names)
	for name := range names {
		if !found[name] {
			t.Errorf("%s is named as pinning an entry but is no test of the repository", name)
		}
	}
}

// newChangeEnv starts a core, signs the owner in and sets up what the drivers
// call: the demo's registry and inputs, a broker account the connector
// reported, an empty spare account, a suggested transfer and a flagged
// balance assertion with their attention items, and flagged price checks.
func newChangeEnv(t *testing.T) *changeEnv {
	t.Helper()
	dsn := emptyDatabase(t)
	code := enrollCode(t, enrollCmd(t, dsn))
	c, srv, pool := startCoreForUI(t, dsn, func(*serveConfig) {})
	ac := authClient(srv.URL, srv.Client(), uiCaller)
	e := &changeEnv{t: t, srv: srv, pool: pool, session: passkeyLogin(t, ac, passkeyEnroll(t, ac, code))}
	e.withClients()
	saveDemoRegistry(t, dsn)
	ds := recordDemoInputs(t, pool)
	e.bondISIN = ds.Bonds[0].Instrument.ISIN.String()
	storeCheckedPrices(t, pool)
	port := serveConnector(t, brokerConnector{})
	if _, err := connectorCmd(t, dsn, fakeCluster(t, port, "ibkr-main"), "sync", "--instance=ibkr-main", "--mode=transactions"); err != nil {
		t.Fatalf("sync of the broker connector: %v", err)
	}
	if _, err := e.accounts.CreateAccount(t.Context(), connect.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: localUI, SessionToken: e.session, Name: "Spare", Sleeve: "satellite", Institution: "bank",
	})); err != nil {
		t.Fatalf("CreateAccount of the spare account: %v", err)
	}
	l := ledgerapp.New(ledgerpg.New(pool))
	pair := importSuggestedPair(t, l)
	e.pair = &ledgerv1.ConfirmTransferRequest{
		Origin: localUI, SessionToken: e.session,
		Withdrawal: &ledgerv1.TransactionKey{Source: pair.Withdrawal.Source, ExternalId: pair.Withdrawal.ExternalID},
		Deposit:    &ledgerv1.TransactionKey{Source: pair.Deposit.Source, ExternalId: pair.Deposit.ExternalID},
	}
	importFlaggedAssertion(t, l)
	c.attention.tick(t.Context(), quiet())
	var items viewsv1.ListAttentionItemsResponse
	getJSON(t, srv, "/v1/attention-items", e.session, &items)
	for _, it := range items.GetItems() {
		switch it.GetKind() {
		case "transfer-suggested":
			e.suggestedItem = it.GetId()
		case "balance-assertion-flagged":
			e.flaggedItem = it.GetId()
		}
	}
	if e.suggestedItem == 0 || e.flaggedItem == 0 {
		t.Fatalf("attention items = %v, want a suggested transfer and a flagged assertion", items.GetItems())
	}
	return e
}

// recordDemoInputs records the demo dataset's valuation inputs, as loadDemo
// does, after the registry: the bonds' terms, prices and FX rates.
func recordDemoInputs(t *testing.T, pool *pgxpool.Pool) *demo.Dataset {
	t.Helper()
	today, err := kernel.ParseDate("2026-10-28")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	ds, err := demo.Generate(demoSeed, today)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	vstore, err := valuationpg.NewStore(pool)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	store := ledgerpg.New(pool)
	l := ledgerapp.New(store, vstore.ImportHook())
	recordInputs(t, valuationapp.NewInputRecorder(valuationapp.NewPlanner(l, store), vstore, jurisdictionPack()), &ds)
	return &ds
}

// entriesAfter returns the audit entries after seq, oldest first.
func entriesAfter(t *testing.T, pool *pgxpool.Pool, seq int64) []identity.AuditEntry {
	t.Helper()
	drainAudit(t, pool)
	page, err := identitypg.New(pool).AuditEntries(t.Context(), &identity.AuditFilter{Limit: identity.MaxAuditPage})
	if err != nil {
		t.Fatalf("AuditEntries: %v", err)
	}
	var out []identity.AuditEntry
	for _, e := range slices.Backward(page.Entries) {
		if e.Seq > seq {
			out = append(out, e)
		}
	}
	return out
}

// wantEntriesOf fails t unless the audit log holds, oldest first, exactly want
// among its entries of the actions in kinds: the changes the owner made and
// the log kept, none for a call that was refused or failed (D-211).
func wantEntriesOf(t *testing.T, pool *pgxpool.Pool, kinds []identity.AuditAction, want ...identity.AuditAction) {
	t.Helper()
	var got []identity.AuditAction
	for _, e := range entriesAfter(t, pool, 0) {
		if slices.Contains(kinds, e.Action) {
			got = append(got, e.Action)
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("the audit log's entries of %v = %v, want %v", kinds, got, want)
	}
}

// lastSeq returns the sequence number of the newest audit entry.
func lastSeq(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	drainAudit(t, pool)
	page, err := identitypg.New(pool).AuditEntries(t.Context(), &identity.AuditFilter{Limit: 1})
	if err != nil {
		t.Fatalf("AuditEntries: %v", err)
	}
	if len(page.Entries) == 0 {
		return 0
	}
	return page.Entries[0].Seq
}

// TestEveryOwnerChangeWritesExactlyOneEntry drives every procedure that
// changes something for the owner, through the real core, and checks that it
// leaves exactly one audit entry, ok, of the owner's session and of the right
// action (D-211, GR-12); then takes the audit log away and checks that no
// change succeeds without it: a change whose entry can't be written is not
// made, and the idempotent ones answer unavailable.
func TestEveryOwnerChangeWritesExactlyOneEntry(t *testing.T) {
	t.Parallel()
	e := newChangeEnv(t)
	var owner string
	for _, c := range ownerChanges() {
		before := lastSeq(t, e.pool)
		if err := c.drive(e); err != nil {
			t.Fatalf("%s: %v", c.procedure, err)
		}
		owner = checkOneEntry(t, &c, entriesAfter(t, e.pool, before), owner)
	}
	// The entry is enqueued in the change's own transaction, so with the
	// outbox gone nothing can be changed: the enqueue fails, the change rolls
	// back and the owner is answered as unavailable.
	if _, err := e.pool.Exec(t.Context(), "ALTER TABLE river.river_job ADD CONSTRAINT no_audit_jobs CHECK (kind <> 'identity.audit_entry') NOT VALID"); err != nil {
		t.Fatalf("taking the outbox away: %v", err)
	}
	for _, c := range ownerChanges() {
		checkRefusedWithoutTheOutbox(t, &c, c.drive(e))
	}
}

// TestOwnerChangesWaitForTheLogInTheOutbox checks that a log that is down does
// not stop the owner's changes, and loses no entry: the entry waits in the
// outbox, and the worker writes it, once, when the log is back, with the time
// of the change (D-211).
func TestOwnerChangesWaitForTheLogInTheOutbox(t *testing.T) {
	t.Parallel()
	e := newChangeEnv(t)
	drainAudit(t, e.pool)
	if _, err := e.pool.Exec(t.Context(), "ALTER TABLE identity.audit_log RENAME TO audit_log_away"); err != nil {
		t.Fatalf("taking the audit log away: %v", err)
	}
	if err := ownerChanges()[0].drive(e); err != nil {
		t.Fatalf("a change with the log down: %v, want it made: its entry waits in the outbox", err)
	}
	// The core's River client retries the job meanwhile, as it would in
	// production, and fails while the log is gone.
	id, args := waitingSettingsEntry(t, e.pool)
	w := &auditWorker{log: identitypg.New(e.pool)}
	if err := w.write(t.Context(), id, &args); err == nil {
		t.Fatal("the worker wrote an entry with the log gone")
	}
	if _, err := e.pool.Exec(t.Context(), "ALTER TABLE identity.audit_log_away RENAME TO audit_log"); err != nil {
		t.Fatalf("bringing the log back: %v", err)
	}
	for range 2 {
		if err := w.write(t.Context(), id, &args); err != nil {
			t.Fatalf("the worker with the log back: %v", err)
		}
	}
	var n int
	var at time.Time
	if err := e.pool.QueryRow(t.Context(), `SELECT count(*), max(at) FROM identity.audit_log WHERE outbox_job_id = $1`, id).Scan(&n, &at); err != nil || n != 1 {
		t.Fatalf("entries of the job = %d, %v; want one even though the worker ran twice", n, err)
	}
	if !at.Equal(args.At) {
		t.Errorf("the entry's time = %s, want the change's, %s: not the worker's", at, args.At)
	}
}

// waitingSettingsEntry returns the ID and the entry of the newest settings
// change the outbox holds.
func waitingSettingsEntry(t *testing.T, pool *pgxpool.Pool) (int64, auditEntryArgs) {
	t.Helper()
	var (
		id  int64
		raw []byte
	)
	if err := pool.QueryRow(t.Context(), `
		SELECT id, args FROM river.river_job WHERE kind = $1 AND args->>'action' = 'settings_changed' ORDER BY id DESC LIMIT 1`,
		auditEntryArgs{}.Kind()).Scan(&id, &raw); err != nil {
		t.Fatalf("reading the entry waiting in the outbox: %v", err)
	}
	var args auditEntryArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		t.Fatalf("decoding the job: %v", err)
	}
	return id, args
}

// checkOneEntry fails t unless got is the one ok entry of c's action by the
// owner, naming the session owner when it is not empty, and returns the
// session the entry names.
func checkOneEntry(t *testing.T, c *ownerChange, got []identity.AuditEntry, owner string) string {
	t.Helper()
	if len(got) != 1 || got[0].Action != c.action || got[0].Outcome != identity.OutcomeOK || got[0].Actor.Kind != identity.ActorOwner {
		t.Errorf("%s left %+v, want exactly one ok entry of %s by the owner", c.procedure, got, c.action)
		return owner
	}
	if owner != "" && got[0].Actor.ID != owner {
		t.Errorf("%s: the entry names the session %q, want the session that signed in, %q", c.procedure, got[0].Actor.ID, owner)
	}
	return got[0].Actor.ID
}

// checkRefusedWithoutTheOutbox fails t when err, c's answer with the outbox
// gone, is a success, or, for an idempotent change, is not unavailable.
func checkRefusedWithoutTheOutbox(t *testing.T, c *ownerChange, err error) {
	t.Helper()
	switch {
	case err == nil:
		t.Errorf("%s succeeded with the outbox gone: a change was made without its entry", c.procedure)
	case c.idempotent && connect.CodeOf(err) != connect.CodeUnavailable:
		t.Errorf("%s with the outbox gone = %v, want unavailable", c.procedure, err)
	}
}

// importFlaggedAssertion imports a balance assertion of 100 PLN cash in
// acc-out, which holds nothing but a withdrawal, so it is flagged and the
// checks raise an item the owner may resolve.
func importFlaggedAssertion(t *testing.T, l *ledgerapp.Ledger) {
	t.Helper()
	day, dayErr := kernel.ParseDate("2026-10-25")
	pln, plnErr := kernel.ParseCurrency("PLN")
	if dayErr != nil || plnErr != nil {
		t.Fatalf("ParseDate, ParseCurrency: %v, %v", dayErr, plnErr)
	}
	a := ledger.CashAssertion("acc-out", day, kernel.NewMoney(kernel.NewDecimal(100, 0), pln))
	if _, err := l.Import(t.Context(), nil, []ledger.BalanceAssertion{a}); err != nil {
		t.Fatalf("Import of the assertion: %v", err)
	}
}

package connectapi_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	"github.com/arkadiuss/openfief/internal/modelrouter/connectapi"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// usageOf is a UsageStore whose day total is fixed.
type usageOf struct{ spent kernel.Decimal }

func (u usageOf) Spent(context.Context, kernel.Date, kernel.Currency) (kernel.Decimal, error) {
	return u.spent, nil
}
func (usageOf) Record(context.Context, *app.UsageRecord) error { return nil }

// questionsOf is a QuestionStore that rules as it is told, and counts what it
// was asked.
type questionsOf struct {
	mu     sync.Mutex
	ruling app.QuestionRuling
	asked  []app.QuestionAsked
}

func (q *questionsOf) Admit(_ context.Context, a *app.QuestionAsked) (app.QuestionRuling, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.asked = append(q.asked, *a)
	return q.ruling, nil
}

// slowProvider answers after it is released, and says when a call began.
type slowProvider struct {
	*provider

	started chan struct{}
	release chan struct{}
}

func (p *slowProvider) Generate(ctx context.Context, a, b string, req *modelrouter.Request) (modelrouter.Result, error) {
	p.started <- struct{}{}
	<-p.release
	return p.provider.Generate(ctx, a, b, req)
}

// guardedAsk is the AskModelService over a router with the guards given and an
// Ask door that counts questions in questions.
type guardedAsk struct {
	client    modelsv1connect.AskModelServiceClient
	provider  *slowProvider
	audit     *audit
	questions *questionsOf
}

func newGuardedAsk(t *testing.T, g app.Guards, limits modelrouter.QuestionLimits) *guardedAsk {
	t.Helper()
	m := modelrouter.Ref{Instance: "p", Model: "m"}
	store := &memStore{s: modelrouter.Settings{
		Switch: modelrouter.Switch{Enabled: true, Toggles: map[modelrouter.Feature]bool{modelrouter.FeatureAsk: true}},
		Routes: map[modelrouter.Role]modelrouter.Route{modelrouter.RoleAsk: {Role: modelrouter.RoleAsk, Models: []modelrouter.Ref{m}}},
		Policy: modelrouter.DataTierPolicy{modelrouter.DataTierT2: {m}, modelrouter.DataTierT0: {m}},
	}}
	prov := &slowProvider{
		provider: &provider{reply: modelrouter.Result{Content: "the answer", Finish: modelrouter.FinishStop}},
		started:  make(chan struct{}, 4), release: make(chan struct{}, 4),
	}
	aud := &audit{}
	var mu sync.Mutex
	n := 0
	router := app.NewRouter(&app.RouterPorts{
		Settings: store, Providers: prov, Audit: aud, Attention: noAttention{},
		CallIDs: func() (string, error) {
			mu.Lock()
			defer mu.Unlock()
			n++
			return fmt.Sprintf("%016x", n), nil
		},
		Now:    func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
		Guards: g,
	})
	qs := &questionsOf{ruling: app.QuestionRuling{Allowed: true}}
	ask := app.NewAsk(askers{}, store, router).WithQuestionLimits(qs, limits)
	path, h := connectapi.NewAsk(ask, passThrough{}, slog.New(slog.NewTextHandler(&strings.Builder{}, nil)))
	mux := http.NewServeMux()
	mux.Handle(path, h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &guardedAsk{client: modelsv1connect.NewAskModelServiceClient(srv.Client(), srv.URL), provider: prov, audit: aud, questions: qs}
}

func (f *guardedAsk) admit(token, address string) (*modelsv1.AdmitAskQuestionResponse, error) {
	resp, err := f.client.AdmitAskQuestion(context.Background(), connect.NewRequest(&modelsv1.AdmitAskQuestionRequest{
		Origin: "o", SessionToken: token, ClientAddress: address,
	}))
	if err != nil {
		return nil, fmt.Errorf("AdmitAskQuestion: %w", err)
	}
	return resp.Msg, nil
}

func (f *guardedAsk) generate(token string) error {
	_, err := f.client.GenerateAsk(context.Background(), connect.NewRequest(&modelsv1.GenerateAskRequest{
		Origin: "o", SessionToken: token, Messages: []*modelsv1.AskMessage{userMessage("T2")}, MaxOutputTokens: 800,
	}))
	if err != nil {
		return fmt.Errorf("GenerateAsk: %w", err)
	}
	return nil
}

func budgetOf(t *testing.T, limit string) modelrouter.Budget {
	t.Helper()
	capacity, err := kernel.ParseDecimal(limit)
	if err != nil {
		t.Fatal(err)
	}
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatal(err)
	}
	b, err := modelrouter.NewBudget(capacity, usd)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAdmitAskQuestionByDecision(t *testing.T) {
	t.Parallel()
	resets := time.Date(2026, 1, 2, 4, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		ruling  app.QuestionRuling
		spent   string
		token   string
		want    *modelsv1.AdmitAskQuestionResponse
		counted int // questions the store was asked to count
	}{
		{
			name: "allowed", ruling: app.QuestionRuling{Allowed: true}, spent: "0", token: "owner", counted: 1,
			want: &modelsv1.AdmitAskQuestionResponse{Admission: modelsv1.AskAdmission_ASK_ADMISSION_ALLOWED},
		},
		{
			name: "the session's limit", ruling: app.QuestionRuling{Subject: app.SubjectSession, ResetsAt: resets}, spent: "0", token: "owner", counted: 1,
			want: &modelsv1.AdmitAskQuestionResponse{Admission: modelsv1.AskAdmission_ASK_ADMISSION_SESSION_LIMIT, ResetsAt: timestamppb.New(resets)},
		},
		{
			name: "the address's limit", ruling: app.QuestionRuling{Subject: app.SubjectAddress, ResetsAt: resets}, spent: "0", token: "owner", counted: 1,
			want: &modelsv1.AdmitAskQuestionResponse{Admission: modelsv1.AskAdmission_ASK_ADMISSION_ADDRESS_LIMIT, ResetsAt: timestamppb.New(resets)},
		},
		{
			name: "the budget is spent, which counts nothing", ruling: app.QuestionRuling{Allowed: true}, spent: "5", token: "owner",
			want: &modelsv1.AdmitAskQuestionResponse{Admission: modelsv1.AskAdmission_ASK_ADMISSION_BUDGET_SPENT, ResetsAt: timestamppb.New(tomorrow)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spent, err := kernel.ParseDecimal(tt.spent)
			if err != nil {
				t.Fatal(err)
			}
			f := newGuardedAsk(t, app.Guards{Budget: budgetOf(t, "5"), Usage: usageOf{spent: spent}}, modelrouter.QuestionLimits{PerHour: 10})
			f.questions.ruling = tt.ruling
			got, err := f.admit(tt.token, "198.51.100.7")
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("response (-want +got):\n%s", diff)
			}
			asked := f.questions.asked
			if len(asked) != tt.counted || (tt.counted == 1 && asked[0].Address != "198.51.100.7") {
				t.Errorf("questions counted = %+v, want %d, under the address the plugin passed", asked, tt.counted)
			}
		})
	}
}

func TestAdmitAskQuestionOfNoSessionIsUnauthenticated(t *testing.T) {
	t.Parallel()
	f := newGuardedAsk(t, app.Guards{}, modelrouter.QuestionLimits{PerHour: 10})
	if _, err := f.admit("nobody", "198.51.100.7"); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("code = %v, want UNAUTHENTICATED (%v)", connect.CodeOf(err), err)
	}
	if len(f.questions.asked) != 0 {
		t.Errorf("a question of no session was counted: %+v", f.questions.asked)
	}
}

func TestGenerateAskRefusedByTheBudgetSaysSoAndWhen(t *testing.T) {
	t.Parallel()
	spent, err := kernel.ParseDecimal("5")
	if err != nil {
		t.Fatal(err)
	}
	f := newGuardedAsk(t, app.Guards{Budget: budgetOf(t, "5"), Usage: usageOf{spent: spent}}, modelrouter.QuestionLimits{})
	err = f.generate("owner")
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeResourceExhausted {
		t.Fatalf("error = %v, want RESOURCE_EXHAUSTED", err)
	}
	if got := ce.Meta().Get(connectapi.RefusalHeader); got != connectapi.RefusalBudget {
		t.Errorf("%s = %q, want %q", connectapi.RefusalHeader, got, connectapi.RefusalBudget)
	}
	if got := ce.Meta().Get(connectapi.ResetsAtHeader); got != "2026-01-03T00:00:00Z" {
		t.Errorf("%s = %q, want the start of the next day", connectapi.ResetsAtHeader, got)
	}
	if strings.Contains(ce.Error(), secretText) {
		t.Errorf("the refusal quotes the question: %v", ce)
	}
	f.audit.mu.Lock()
	defer f.audit.mu.Unlock()
	if len(f.audit.results) != 1 || f.audit.results[0].Class != app.ClassBudget || f.audit.results[0].Outcome != app.OutcomeRefused {
		t.Errorf("audit = %+v, want one refusal of class budget", f.audit.results)
	}
}

func TestGenerateAskRefusedAsBusyWhenTheCallsAtOnceAreTaken(t *testing.T) {
	t.Parallel()
	f := newGuardedAsk(t, app.Guards{AskCalls: 1, QueueWait: 10 * time.Millisecond}, modelrouter.QuestionLimits{})
	first := make(chan error, 1)
	go func() { first <- f.generate("owner") }()
	<-f.provider.started
	err := f.generate("owner")
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeResourceExhausted || ce.Meta().Get(connectapi.RefusalHeader) != connectapi.RefusalBusy {
		t.Fatalf("error = %v, want RESOURCE_EXHAUSTED refusing as busy", err)
	}
	f.provider.release <- struct{}{}
	if err := <-first; err != nil {
		t.Errorf("the call within the limit: %v", err)
	}
}

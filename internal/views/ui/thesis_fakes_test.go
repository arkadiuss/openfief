package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	thesisv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/thesis/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/thesis/v1/thesisv1connect"
)

// fakeTheses is an in-memory ThesisService: it wants the plugin's token and the
// owner's session, keeps every version, and refuses what the core would (a
// draft with no statement or no horizon, a stale edit), recording each change.
type fakeTheses struct {
	thesisv1connect.UnimplementedThesisServiceHandler

	mu    sync.Mutex
	order []string
	byID  map[string][]*thesisv1.ThesisVersion
	// calls lists the changes made, as "save <id>" and "edit <id> v<base>".
	calls []string
	clock time.Time
}

func newFakeTheses() *fakeTheses {
	return &fakeTheses{byID: map[string][]*thesisv1.ThesisVersion{}, clock: time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)}
}

var (
	errNeedsStatement = errors.New("invalid thesis: the statement is 1 to 1000 characters")
	errHorizonWords   = errors.New("invalid thesis: the horizon is long_term or short_term")
	errListEmpty      = errors.New("invalid thesis: a list is empty")
	errNewerVersion   = errors.New("the thesis has a newer version")
)

// check refuses a draft as the core does, for the cases the tests make.
func check(d *thesisv1.ThesisDraft) error {
	switch {
	case strings.TrimSpace(d.GetStatement()) == "":
		return codeErr(connect.CodeInvalidArgument, errNeedsStatement)
	case d.GetHorizon() == thesisv1.Horizon_HORIZON_UNSPECIFIED:
		return codeErr(connect.CodeInvalidArgument, errHorizonWords)
	case len(d.GetInstruments()) == 0 || len(d.GetInvalidation()) == 0 || len(d.GetSources()) == 0:
		return codeErr(connect.CodeInvalidArgument, errListEmpty)
	}
	return nil
}

func (f *fakeTheses) SaveThesis(_ context.Context, r *connect.Request[thesisv1.SaveThesisRequest]) (*connect.Response[thesisv1.SaveThesisResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if err := check(r.Msg.GetDraft()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	id := fmt.Sprintf("%016x", len(f.order)+1)
	f.order = append(f.order, id)
	f.byID[id] = []*thesisv1.ThesisVersion{f.version(1, r.Msg.GetDraft())}
	f.calls = append(f.calls, "save "+id)
	return connect.NewResponse(&thesisv1.SaveThesisResponse{Thesis: f.thesis(id)}), nil
}

func (f *fakeTheses) EditThesis(_ context.Context, r *connect.Request[thesisv1.EditThesisRequest]) (*connect.Response[thesisv1.EditThesisResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	id := r.Msg.GetThesisId()
	versions := f.byID[id]
	if len(versions) == 0 {
		return nil, codeErr(connect.CodeNotFound, nil)
	}
	if err := check(r.Msg.GetDraft()); err != nil {
		return nil, err
	}
	if int(r.Msg.GetBaseVersion()) != len(versions) {
		return nil, codeErr(connect.CodeFailedPrecondition, errNewerVersion)
	}
	f.byID[id] = append(versions, f.version(len(versions)+1, r.Msg.GetDraft()))
	f.calls = append(f.calls, fmt.Sprintf("edit %s v%d", id, r.Msg.GetBaseVersion()))
	return connect.NewResponse(&thesisv1.EditThesisResponse{Thesis: f.thesis(id)}), nil
}

func (f *fakeTheses) GetThesis(_ context.Context, r *connect.Request[thesisv1.GetThesisRequest]) (*connect.Response[thesisv1.GetThesisResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.byID[r.Msg.GetThesisId()]) == 0 {
		return nil, codeErr(connect.CodeNotFound, nil)
	}
	return connect.NewResponse(&thesisv1.GetThesisResponse{Thesis: f.thesis(r.Msg.GetThesisId())}), nil
}

func (f *fakeTheses) ListTheses(_ context.Context, r *connect.Request[thesisv1.ListThesesRequest]) (*connect.Response[thesisv1.ListThesesResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	out := &thesisv1.ListThesesResponse{}
	for _, id := range slices.Backward(f.order) {
		vs := f.byID[id]
		last := vs[len(vs)-1]
		out.Theses = append(out.Theses, &thesisv1.ThesisSummary{
			Id: id, Version: last.GetVersion(), CreatedAt: vs[0].GetSavedAt(), UpdatedAt: last.GetSavedAt(), Draft: last.GetDraft(), DataTier: "T1",
		})
	}
	return connect.NewResponse(out), nil
}

// admit checks the plugin's token and that the session is the owner's.
func (f *fakeTheses) admit(h http.Header, token string) error {
	if h.Get("Authorization") != "Bearer "+pluginToken {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	if _, err := scopeOf(token); err != nil {
		return err
	}
	return owner(token)
}

// version makes the next version, stamped a minute after the last one.
func (f *fakeTheses) version(n int, d *thesisv1.ThesisDraft) *thesisv1.ThesisVersion {
	f.clock = f.clock.Add(time.Minute)
	return &thesisv1.ThesisVersion{Version: int32(n), SavedAt: timestamppb.New(f.clock), Draft: clonedDraft(d)} //nolint:gosec // a test's small count
}

// clonedDraft returns a copy of d.
func clonedDraft(d *thesisv1.ThesisDraft) *thesisv1.ThesisDraft {
	if c, ok := proto.Clone(d).(*thesisv1.ThesisDraft); ok {
		return c
	}
	return &thesisv1.ThesisDraft{}
}

// thesis returns the thesis id with its versions.
func (f *fakeTheses) thesis(id string) *thesisv1.Thesis {
	return &thesisv1.Thesis{Id: id, Versions: f.byID[id], DataTier: "T1"}
}

// seed saves a draft as the owner's first version, without the calls list.
func (f *fakeTheses) seed(d *thesisv1.ThesisDraft) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := fmt.Sprintf("%016x", len(f.order)+1)
	f.order = append(f.order, id)
	f.byID[id] = []*thesisv1.ThesisVersion{f.version(1, d)}
	return id
}

// seedEdit saves a draft as the next version of thesis id, without the calls list.
func (f *fakeTheses) seedEdit(id string, d *thesisv1.ThesisDraft) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[id] = append(f.byID[id], f.version(len(f.byID[id])+1, d))
}

func (f *fakeTheses) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

// NewSleeves returns the SleeveService's path and HTTP handler over sleeves
// (D-71, D-79), with callers in front of every call to admit view plugins
// only, logging to log.
func NewSleeves(sleeves *app.Sleeves, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return portfoliov1connect.NewSleeveServiceHandler(&sleeveService{sleeves: sleeves, log: log}, connect.WithInterceptors(callers))
}

// sleeveService implements portfoliov1connect.SleeveServiceHandler.
type sleeveService struct {
	sleeves *app.Sleeves
	log     *slog.Logger
}

// AddSleeveKind adds the owner's sleeve kind.
func (s *sleeveService) AddSleeveKind(ctx context.Context, req *connect.Request[portfoliov1.AddSleeveKindRequest]) (*connect.Response[portfoliov1.AddSleeveKindResponse], error) {
	m := req.Msg
	k, err := s.sleeves.AddKind(ctx, m.GetOrigin(), m.GetSessionToken(), m.GetName(), roleFrom(m.GetDefaultRole()))
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.AddSleeveKindResponse{Kind: pbKind(&k)}), nil
}

// RenameSleeveKind renames a sleeve kind.
func (s *sleeveService) RenameSleeveKind(ctx context.Context, req *connect.Request[portfoliov1.RenameSleeveKindRequest]) (*connect.Response[portfoliov1.RenameSleeveKindResponse], error) {
	m := req.Msg
	k, err := s.sleeves.RenameKind(ctx, m.GetOrigin(), m.GetSessionToken(), portfolio.KindKey(m.GetKey()), m.GetName())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.RenameSleeveKindResponse{Kind: pbKind(&k)}), nil
}

// AddSleeve adds the owner's sleeve.
func (s *sleeveService) AddSleeve(ctx context.Context, req *connect.Request[portfoliov1.AddSleeveRequest]) (*connect.Response[portfoliov1.AddSleeveResponse], error) {
	m := req.Msg
	sl, err := s.sleeves.AddSleeve(ctx, m.GetOrigin(), m.GetSessionToken(), portfolio.Sleeve{
		Name: m.GetName(), Kind: portfolio.KindKey(m.GetKind()), Role: roleFrom(m.GetRole()), RoleOverride: m.GetRoleOverride(),
	})
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.AddSleeveResponse{Sleeve: pbSleeve(&sl)}), nil
}

// UpdateSleeve changes a sleeve.
func (s *sleeveService) UpdateSleeve(ctx context.Context, req *connect.Request[portfoliov1.UpdateSleeveRequest]) (*connect.Response[portfoliov1.UpdateSleeveResponse], error) {
	m := req.Msg
	sl, err := s.sleeves.UpdateSleeve(ctx, m.GetOrigin(), m.GetSessionToken(), portfolio.Sleeve{
		ID: portfolio.SleeveID(m.GetId()), Name: m.GetName(), Kind: portfolio.KindKey(m.GetKind()),
		Role: roleFrom(m.GetRole()), RoleOverride: m.GetRoleOverride(),
	})
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.UpdateSleeveResponse{Sleeve: pbSleeve(&sl)}), nil
}

// roleFrom reads a role; unspecified, or a number of no role, is the zero
// Role, which takes the kind's default role or fails the kind's check.
func roleFrom(r portfoliov1.SleeveRole) portfolio.Role {
	switch r {
	case portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE:
		return portfolio.RoleDefensive
	case portfoliov1.SleeveRole_SLEEVE_ROLE_OFFENSIVE:
		return portfolio.RoleOffensive
	case portfoliov1.SleeveRole_SLEEVE_ROLE_UNSPECIFIED:
		return 0
	}
	return 0
}

// pbRole writes a role.
func pbRole(r portfolio.Role) portfoliov1.SleeveRole {
	switch r {
	case portfolio.RoleDefensive:
		return portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE
	case portfolio.RoleOffensive:
		return portfoliov1.SleeveRole_SLEEVE_ROLE_OFFENSIVE
	}
	return portfoliov1.SleeveRole_SLEEVE_ROLE_UNSPECIFIED
}

// pbKind converts a sleeve kind.
func pbKind(k *portfolio.Kind) *portfoliov1.SleeveKind {
	return &portfoliov1.SleeveKind{Key: string(k.Key), Name: k.Name, DefaultRole: pbRole(k.DefaultRole)}
}

// pbSleeve converts a sleeve.
func pbSleeve(sl *portfolio.Sleeve) *portfoliov1.Sleeve {
	return &portfoliov1.Sleeve{
		Id: string(sl.ID), Name: sl.Name, Kind: string(sl.Kind), Role: pbRole(sl.Role), RoleOverride: sl.RoleOverride,
	}
}

// fail maps err to its Connect error. A broken rule carries its text,
// which names the kind or sleeve and the rule, so the owner can fix it.
func (s *sleeveService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, portfolio.ErrInvalidKind), errors.Is(err, portfolio.ErrInvalidSleeve),
		errors.Is(err, portfolio.ErrUnknownKind), errors.Is(err, portfolio.ErrKindRoleChange),
		errors.Is(err, portfolio.ErrRoleOverride), errors.Is(err, portfolio.ErrNoKey):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, app.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "sleeve call failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}

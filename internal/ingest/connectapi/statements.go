package connectapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

const (
	// maxMessageBytes caps one message of an upload stream: a header or one
	// chunk. The built-in UI sends 64 KiB chunks; the whole file is capped by
	// the use case.
	maxMessageBytes = 1 << 20
	// maxProblems is how many ingest problems an answer lists; the rest are
	// only counted.
	maxProblems = 100
)

// Errors of a malformed upload stream.
var (
	// errNoHeader reports a stream whose first message is not a header.
	errNoHeader = errors.New("the upload must start with its header")
	// errSecondHeader reports a header after the first message.
	errSecondHeader = errors.New("the upload has a second header")
	// errReceive reports a stream that broke off.
	errReceive = errors.New("the upload broke off")
)

// NewStatements returns the StatementService's path and HTTP handler over
// uploads, with callers in front of every call to admit view plugins only,
// logging to log. One message may carry at most 1 MiB.
func NewStatements(uploads *app.Uploads, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return ingestv1connect.NewStatementServiceHandler(&statements{uploads: uploads, log: log},
		connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxMessageBytes))
}

// statements implements ingestv1connect.StatementServiceHandler.
type statements struct {
	uploads *app.Uploads
	log     *slog.Logger
}

// UploadStatement reads the header, then hands the rest of the stream to
// the use case as the file, which reads it only once the owner and the
// account's connector are checked.
func (s *statements) UploadStatement(ctx context.Context, stream *connect.ClientStream[ingestv1.UploadStatementRequest]) (*connect.Response[ingestv1.UploadStatementResponse], error) {
	if !stream.Receive() {
		if err := stream.Err(); err != nil {
			return nil, s.fail(ctx, fmt.Errorf("%w: %w", errReceive, err))
		}
		return nil, s.fail(ctx, errNoHeader)
	}
	h := stream.Msg().GetHeader()
	if h == nil {
		return nil, s.fail(ctx, errNoHeader)
	}
	file := app.NamedFile{Name: h.GetFileName(), Contents: &chunkReader{stream: stream}}
	b, err := s.uploads.Upload(ctx, h.GetOrigin(), h.GetSessionToken(), ledger.AccountID(h.GetAccountId()), modeFrom(h.GetMode()), file)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(uploadResponse(&b)), nil
}

// fail maps err to its Connect error. A refusal carries its text, which
// names the rule, the account or the connector's reason, so the owner can
// fix it; it never carries an amount.
func (s *statements) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, errReceive):
		return connect.NewError(receiveCode(err), errReceive)
	case errors.Is(err, app.ErrFileTooLarge):
		return connect.NewError(connect.CodeResourceExhausted, app.ErrFileTooLarge)
	case errors.Is(err, errNoHeader):
		return connect.NewError(connect.CodeInvalidArgument, errNoHeader)
	case errors.Is(err, app.ErrInvalidUpload), errors.Is(err, errSecondHeader), errors.Is(err, app.ErrEntryRefused),
		errors.Is(err, app.ErrSeedingOutsideOpening):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrNoConnector), errors.Is(err, app.ErrNoEntries), errors.Is(err, ingest.ErrAccountNotBound),
		errors.Is(err, ledger.ErrAlreadySeeded):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, app.ErrConnectorUnavailable):
		return connect.NewError(connect.CodeUnavailable, app.ErrConnectorUnavailable)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "statement upload failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}

// receiveCode is the code of a stream that broke off: the stream's own,
// such as resource_exhausted for a message over the limit or canceled for a
// caller that went away, else unknown.
func receiveCode(err error) connect.Code {
	if ce, ok := errors.AsType[*connect.Error](err); ok {
		return ce.Code()
	}
	return connect.CodeUnknown
}

// chunkReader reads the file's bytes from the chunks after the header. A
// second header, or a stream that broke off, fails the read.
type chunkReader struct {
	stream *connect.ClientStream[ingestv1.UploadStatementRequest]
	buf    []byte
}

// Read returns the next bytes of the current chunk, receiving the next one
// when it is used up, and io.EOF after the last.
func (c *chunkReader) Read(p []byte) (int, error) {
	for len(c.buf) == 0 {
		if !c.stream.Receive() {
			if err := c.stream.Err(); err != nil {
				return 0, fmt.Errorf("%w: %w", errReceive, err)
			}
			return 0, io.EOF
		}
		if c.stream.Msg().GetHeader() != nil {
			return 0, errSecondHeader
		}
		c.buf = c.stream.Msg().GetChunk()
	}
	n := copy(p, c.buf)
	c.buf = c.buf[n:]
	return n, nil
}

// modeFrom reads a request's mode; an unspecified or unknown one is the
// zero ingest.Mode, which the use case refuses.
func modeFrom(m ingestv1.Mode) ingest.Mode {
	switch m {
	case ingestv1.Mode_MODE_OPENING:
		return ingest.ModeOpening
	case ingestv1.Mode_MODE_TRANSACTIONS:
		return ingest.ModeTransactions
	case ingestv1.Mode_MODE_UNSPECIFIED:
	}
	return 0
}

// uploadResponse is the answer for batch b: its counts and its first
// maxProblems ingest problems.
func uploadResponse(b *ingest.Batch) *ingestv1.UploadStatementResponse {
	shown := b.Problems[:min(len(b.Problems), maxProblems)]
	problems := make([]*ingestv1.Problem, len(shown))
	for i, p := range shown {
		problems[i] = &ingestv1.Problem{Kind: problemKindTo(p.Kind), Ref: p.Ref, Reason: p.Reason}
	}
	return &ingestv1.UploadStatementResponse{
		BatchId: string(b.ID), Institution: string(b.Institution),
		Transactions: count(b.Transactions), Assertions: count(b.Assertions),
		New: count(b.Imported.New), NewVersions: count(b.Imported.NewVersions), Unchanged: count(b.Imported.Unchanged),
		FailedAssertions: count(b.Imported.FailedAssertions),
		Problems:         problems, ProblemsTotal: count(len(b.Problems)),
	}
}

// problemKindTo is a problem kind's message enum; an invalid kind is
// unspecified.
func problemKindTo(k ingest.ProblemKind) ingestv1.ProblemKind {
	switch k {
	case ingest.ProblemUnsupported:
		return ingestv1.ProblemKind_PROBLEM_KIND_UNSUPPORTED
	case ingest.ProblemSkipped:
		return ingestv1.ProblemKind_PROBLEM_KIND_SKIPPED
	}
	return ingestv1.ProblemKind_PROBLEM_KIND_UNSPECIFIED
}

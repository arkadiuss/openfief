package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
)

const (
	// maxStatementBytes is the largest statement file the core takes,
	// 64 MiB (ingest/app.DefaultMaxUploadBytes).
	maxStatementBytes = 64 << 20
	// maxUploadBody caps the whole multipart body: the file plus room for
	// the mode field and the part headers.
	maxUploadBody = maxStatementBytes + 64<<10
	// uploadChunk is how many bytes of the file one message to the core
	// carries.
	uploadChunk = 64 << 10
	// maxModeBytes caps the mode field.
	maxModeBytes = 32
	// maxCashBytes caps each opening cash field: an amount or a currency.
	maxCashBytes = 64
)

// Errors of a malformed upload form.
var (
	errUploadForm = errors.New("the upload form is malformed")
	// errUploadTooLarge reports a body over maxUploadBody.
	errUploadTooLarge = errors.New("the statement file is larger than 64 MiB")
	// errUploadBroke reports a body that broke off.
	errUploadBroke = errors.New("the upload broke off")
)

// uploadForm is the "Upload statement" form of an account whose bound
// connector takes statement files (D-109): Mode is the mode the core
// chooses for its next upload, which the owner may override. While it is
// opening, the account hasn't started, and the form asks how it starts:
// from the holdings of a current statement with the cash balance the owner
// enters once, or from a statement of its whole history (D-136).
type uploadForm struct {
	Instance string
	Mode     string
}

// Start reports whether the form asks how the account starts.
func (f *uploadForm) Start() bool { return f.Mode == modeName(ingestv1.Mode_MODE_OPENING) }

// uploadForm returns the account's upload form, and false when its bound
// connector takes no statement files, as the core's StatementService says.
func (s *server) uploadForm(r *http.Request, sess session, account string) (*uploadForm, bool, error) {
	resp, err := s.statements.ListStatementAccounts(r.Context(), connect.NewRequest(&ingestv1.ListStatementAccountsRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		return nil, false, fmt.Errorf("ListStatementAccounts: %w", err)
	}
	for _, a := range resp.Msg.GetAccounts() {
		if a.GetAccountId() == account {
			return &uploadForm{Instance: a.GetInstance(), Mode: modeName(a.GetMode())}, true, nil
		}
	}
	return nil, false, nil
}

// modeName is the form's value of a mode: opening or transactions, empty
// for none.
func modeName(m ingestv1.Mode) string {
	switch m {
	case ingestv1.Mode_MODE_OPENING:
		return "opening"
	case ingestv1.Mode_MODE_TRANSACTIONS:
		return "transactions"
	case ingestv1.Mode_MODE_UNSPECIFIED:
	}
	return ""
}

// uploadMode is the contract's mode for the form's value: auto lets the
// core choose (unspecified, D-109); false for any other value.
func uploadMode(value string) (ingestv1.Mode, bool) {
	switch value {
	case "auto":
		return ingestv1.Mode_MODE_UNSPECIFIED, true
	case "opening":
		return ingestv1.Mode_MODE_OPENING, true
	case "transactions":
		return ingestv1.Mode_MODE_TRANSACTIONS, true
	}
	return ingestv1.Mode_MODE_UNSPECIFIED, false
}

// uploadStatement streams the posted statement file to the core
// (UploadStatement, D-49), which hands it to the connector bound to the
// account, and shows the account page with the import's counts and ingest
// problems. The form must send the mode first, then the file; the file is
// read in chunks and never stored. A refusal, such as a file the connector
// can't read, shows the page with the core's reason.
func (s *server) uploadStatement(w http.ResponseWriter, r *http.Request, sess session) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBody)
	mr, err := r.MultipartReader()
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The upload form is malformed.")
		return
	}
	form, file, err := uploadParts(mr)
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The upload form is malformed: send the mode, then one statement file.")
		return
	}
	header := &ingestv1.StatementHeader{
		Origin: sess.Origin, SessionToken: sess.Token, AccountId: r.PathValue("id"), FileName: file.FileName(), Mode: form.mode,
	}
	if form.cash != "" {
		header.OpeningCash, header.OpeningCashCurrency = form.cash, strings.ToUpper(form.currency)
	}
	resp, err := s.sendStatement(r.Context(), header, file)
	if err != nil {
		s.uploadFailed(w, r, sess, err)
		return
	}
	s.renderAccount(w, r, sess, http.StatusOK, accountForm{Upload: resp})
}

// uploadFields are the upload form's fields before the file: the mode and,
// on an account that hasn't started, the opening cash the owner entered
// and its currency, empty for none (D-136).
type uploadFields struct {
	mode           ingestv1.Mode
	cash, currency string
}

// uploadParts reads the form's first part, the mode, and the opening cash
// fields that may follow it, and returns them with the next part, the
// file, ready to read.
func uploadParts(mr *multipart.Reader) (uploadFields, *multipart.Part, error) {
	var f uploadFields
	part, err := mr.NextPart()
	if err != nil || part.FormName() != "mode" {
		return f, nil, errUploadForm
	}
	value, err := io.ReadAll(io.LimitReader(part, maxModeBytes))
	if err != nil {
		return f, nil, errUploadForm
	}
	mode, ok := uploadMode(string(value))
	if !ok {
		return f, nil, errUploadForm
	}
	f.mode = mode
	file, err := cashParts(mr, &f)
	if err != nil {
		return f, nil, err
	}
	return f, file, nil
}

// cashParts reads into f the opening cash fields that follow the mode, in
// any order, and returns the part after them, which must be the file.
func cashParts(mr *multipart.Reader, f *uploadFields) (*multipart.Part, error) {
	for {
		part, err := mr.NextPart()
		if err != nil {
			return nil, errUploadForm
		}
		fields := map[string]*string{"opening_cash": &f.cash, "opening_cash_currency": &f.currency}
		field, ok := fields[part.FormName()]
		if !ok {
			if part.FormName() != "statement" || part.FileName() == "" {
				return nil, errUploadForm
			}
			return part, nil
		}
		v, err := io.ReadAll(io.LimitReader(part, maxCashBytes))
		if err != nil {
			return nil, errUploadForm
		}
		*field = strings.TrimSpace(string(v))
	}
}

// sendStatement streams header and then file to the core, and returns the
// core's answer or its error. A file that can't be read to its end cancels
// the call without ending the stream, so the core imports nothing of it,
// and is errUploadTooLarge over the body's limit, else errUploadBroke.
func (s *server) sendStatement(ctx context.Context, header *ingestv1.StatementHeader, file io.Reader) (*ingestv1.UploadStatementResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream := s.statements.UploadStatement(ctx)
	sendErr := stream.Send(&ingestv1.UploadStatementRequest{Part: &ingestv1.UploadStatementRequest_Header{Header: header}})
	buf := make([]byte, uploadChunk)
	for sendErr == nil {
		n, err := file.Read(buf)
		if n > 0 {
			chunk := append([]byte(nil), buf[:n]...)
			sendErr = stream.Send(&ingestv1.UploadStatementRequest{Part: &ingestv1.UploadStatementRequest_Chunk{Chunk: chunk}})
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			// Cancel without closing the stream: closing it would end the
			// request body cleanly, and the core could take the truncated
			// file for the whole one.
			cancel()
			if _, tooLarge := errors.AsType[*http.MaxBytesError](err); tooLarge {
				return nil, errUploadTooLarge
			}
			return nil, errUploadBroke
		}
	}
	resp, err := stream.CloseAndReceive()
	if err != nil {
		return nil, fmt.Errorf("UploadStatement: %w", err)
	}
	return resp.Msg, nil
}

// uploadFailed answers a failed upload: a body that broke off as a bad
// request; the core's refusal as the account page again with its reason, a
// connector that can't be reached or a file over the limit with their own
// message; and anything else as any failed call to the core.
func (s *server) uploadFailed(w http.ResponseWriter, r *http.Request, sess session, err error) {
	if errors.Is(err, errUploadBroke) {
		s.errorPage(w, r, http.StatusBadRequest, "The upload broke off. Try again.")
		return
	}
	code := connect.CodeOf(err)
	if errors.Is(err, errUploadTooLarge) {
		code = connect.CodeResourceExhausted
	}
	if code == connect.CodeUnavailable {
		s.renderAccount(w, r, sess, http.StatusServiceUnavailable, accountForm{
			Error: "The statement was not imported: its connector can't be reached now. Try again in a moment.",
		})
		return
	}
	if code == connect.CodeResourceExhausted {
		s.renderAccount(w, r, sess, http.StatusRequestEntityTooLarge, accountForm{Error: "The statement was not imported: the file is larger than 64 MiB."})
		return
	}
	if status, refused := refusalStatus(code); refused && code != connect.CodePermissionDenied {
		s.renderAccount(w, r, sess, status, accountForm{Error: "The statement was not imported: " + connectMessage(err)})
		return
	}
	s.coreFailed(w, r, "UploadStatement", err)
}

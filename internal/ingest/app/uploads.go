package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// DefaultMaxUploadBytes caps a statement file the owner uploads when
// Uploads.MaxBytes is zero: 64 MiB, the largest statement a connector takes
// today (the IBKR Flex parser's limit).
const DefaultMaxUploadBytes int64 = 64 << 20

// sniffBytes is how many of a file's first bytes Uploads.Sniff sees, as
// many as content sniffing reads (http.DetectContentType).
const sniffBytes = 512

// Errors Uploads.Upload wraps for a file the owner may fix.
var (
	// ErrInvalidUpload reports an upload without a mode or with an empty
	// file.
	ErrInvalidUpload = errors.New("invalid statement upload")
	// ErrFileTooLarge reports a file over Uploads.MaxBytes.
	ErrFileTooLarge = errors.New("the statement file is too large")
)

// Uploads takes statement files the owner uploads for an account in the
// built-in UI (D-7) and hands each to the connector instance bound to the
// account, which maps it; the core imports what it delivers like any other
// batch (D-49). The connector decides which files it takes. The file is
// streamed through, never stored.
type Uploads struct {
	Owners     Owners
	Bindings   AccountBindings
	Connectors *Connectors
	// Batches records each completed batch in the import batch history.
	Batches BatchRecorder
	// NewBatchID returns a fresh import batch ID.
	NewBatchID func() ingest.BatchID
	// MaxBytes caps a file; zero is DefaultMaxUploadBytes.
	MaxBytes int64
	// Sniff returns the media type of a file from its first bytes (at most
	// 512), which the connector gets as a hint in place of whatever the
	// owner's browser declared. Nil sends none.
	Sniff func(head []byte) string
}

// Upload checks that the caller is the owner (an owner-scoped session on
// origin), then streams file to the one connector instance bound to
// account as a new import batch in mode, and returns the completed batch,
// after recording it in the import batch history. Only the last element of
// the file's name, cleaned by ingest.FileNameHint, reaches the connector,
// and its media type is Sniff's. Errors wrap ErrNoSession or ErrNotOwner
// for a caller who may not, ErrInvalidUpload for no mode or an empty file,
// ErrFileTooLarge for a file over the limit, ErrNoConnector for an account
// without exactly one bound instance, ErrEntryRefused or ErrNoEntries for a
// connector that won't take the file, ErrConnectorUnavailable for one that
// can't be reached, ingest.ErrAccountNotBound for a delivery into another
// account, and the ledger's ledger.ErrAlreadySeeded or
// ErrSeedingOutsideOpening for a statement in the wrong mode. Uploading the
// same statement again imports nothing new (D-18).
func (u *Uploads) Upload(ctx context.Context, origin, sessionToken string, account ledger.AccountID, mode ingest.Mode, file NamedFile) (ingest.Batch, error) {
	if err := u.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	if mode != ingest.ModeOpening && mode != ingest.ModeTransactions {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w: no mode", ErrInvalidUpload)
	}
	instance, err := boundInstance(ctx, u.Bindings, account)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	out, err := u.prepare(file)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	batch, err := u.Connectors.ImportFile(ctx, instance, u.NewBatchID(), account, mode, out)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	if err := u.Batches.RecordBatch(ctx, instance, &batch); err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	return batch, nil
}

// prepare returns the file as the connector gets it: its name as a hint,
// its media type sniffed from its first bytes, and its contents capped at
// the limit. It reads those first bytes, and refuses an empty file.
func (u *Uploads) prepare(file NamedFile) (NamedFile, error) {
	contents := &cappedReader{r: file.Contents, left: u.maxBytes()}
	head := make([]byte, sniffBytes)
	n, err := io.ReadFull(contents, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return NamedFile{}, fmt.Errorf("the first bytes: %w", err)
	}
	if n == 0 {
		return NamedFile{}, fmt.Errorf("%w: the file is empty", ErrInvalidUpload)
	}
	head = head[:n]
	out := NamedFile{Name: ingest.FileNameHint(file.Name), Contents: io.MultiReader(bytes.NewReader(head), contents)}
	if u.Sniff != nil {
		out.MediaType = u.Sniff(head)
	}
	return out, nil
}

// maxBytes is MaxBytes, or its default.
func (u *Uploads) maxBytes() int64 {
	if u.MaxBytes > 0 {
		return u.MaxBytes
	}
	return DefaultMaxUploadBytes
}

// cappedReader reads r until left bytes are used up; a byte beyond them
// fails the read with ErrFileTooLarge and is never returned.
type cappedReader struct {
	r    io.Reader
	left int64
}

// Read reads from r, refusing to go past the cap. io.EOF is returned as
// is, as readers must.
func (c *cappedReader) Read(p []byte) (int, error) {
	if c.left < 0 {
		return 0, ErrFileTooLarge
	}
	if int64(len(p)) > c.left+1 {
		p = p[:c.left+1]
	}
	n, err := c.r.Read(p)
	if int64(n) > c.left {
		c.left = -1
		return 0, ErrFileTooLarge
	}
	c.left -= int64(n)
	switch {
	case err == nil:
		return n, nil
	case errors.Is(err, io.EOF):
		return n, io.EOF
	}
	return n, fmt.Errorf("reading the file: %w", err)
}

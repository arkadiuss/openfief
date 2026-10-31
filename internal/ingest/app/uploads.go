package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

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
	// ErrInvalidUpload reports an upload with an empty file.
	ErrInvalidUpload = errors.New("invalid statement upload")
	// ErrFileTooLarge reports a file over Uploads.MaxBytes.
	ErrFileTooLarge = errors.New("the statement file is too large")
)

// Uploads takes statement files the owner uploads for an account in the
// built-in UI (D-7) and hands each to the connector instance bound to the
// account, which maps it; the core imports what it delivers like any other
// batch (D-49), every mapped account bound to the instance included
// (D-109). The connector decides which files it takes. The file is
// streamed through, never stored.
type Uploads struct {
	Owners   Owners
	Bindings AccountBindings
	// Connectors imports the delivery. Its Bindings, Declarations and
	// Seeding also tell which accounts take files and in which mode.
	Connectors *Connectors
	// Instances lists the connector instances that run now; nil lists none,
	// so no account takes files.
	Instances InstanceLister
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

// FileAccount is an account the owner can upload statements for: its one
// bound connector instance declares that it takes statement files (D-109),
// and Mode is the mode the core chooses for its next upload.
type FileAccount struct {
	Account  ledger.AccountID
	Instance string
	Mode     ingest.Mode
}

// FileAccounts returns, to the owner, the accounts whose one bound
// connector instance runs and declares statement files in its manifest,
// in account order, each with the mode its next upload takes (D-109). An
// account bound to more than one instance is left out: an upload for it is
// refused.
func (u *Uploads) FileAccounts(ctx context.Context, origin, sessionToken string) ([]FileAccount, error) {
	if err := u.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return nil, fmt.Errorf("listing accounts that take files: %w", err)
	}
	if u.Instances == nil || u.Connectors.Declarations == nil {
		return nil, nil
	}
	instances, err := u.Instances.ConnectorInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing accounts that take files: %w", err)
	}
	var out []FileAccount
	for _, instance := range instances {
		accounts, err := u.fileAccountsOf(ctx, instance)
		if err != nil {
			return nil, fmt.Errorf("listing accounts that take files: %w", err)
		}
		out = append(out, accounts...)
	}
	slices.SortFunc(out, func(a, b FileAccount) int { return strings.Compare(string(a.Account), string(b.Account)) })
	return out, nil
}

// Upload checks that the caller is the owner (an owner-scoped session on
// origin), then streams file to the one connector instance bound to
// account as a new import batch, and returns the completed batch, after
// recording it in the import batch history. The core chooses the mode:
// opening for an account not yet seeded, transactions after that, unless
// mode, the owner's override, is one (D-109, ingest.UploadMode). An
// opening upload seeds the account, so scheduled syncs continue it in
// transactions mode (D-95). A statement of several accounts books every
// mapped account bound to the instance and lists the broker accounts that
// aren't mapped (Batch.Unmapped). Only the last element of the file's
// name, cleaned by ingest.FileNameHint, reaches the connector, and its
// media type is Sniff's. Errors wrap ErrNoSession or ErrNotOwner for a
// caller who may not, ErrInvalidUpload for an empty file,
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
	instance, err := boundInstance(ctx, u.Bindings, account)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	goLive, err := u.Connectors.GoLiveDates(ctx, []ledger.AccountID{account})
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	_, seeded := goLive[account]
	mode = ingest.UploadMode(seeded, mode)
	out, err := u.prepare(file)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	batch, err := u.Connectors.ImportFile(ctx, instance, u.NewBatchID(), mode, out)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	if err := u.Batches.RecordBatch(ctx, instance, &batch); err != nil {
		return ingest.Batch{}, fmt.Errorf("uploading a statement: %w", err)
	}
	return batch, nil
}

// fileAccountsOf returns the accounts bound to instance alone, with their
// next upload's mode, when instance declares statement files.
func (u *Uploads) fileAccountsOf(ctx context.Context, instance string) ([]FileAccount, error) {
	declared, err := u.Connectors.Declarations.Declarations(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("declarations of %s: %w", instance, err)
	}
	if !declared.StatementFiles {
		return nil, nil
	}
	bound, err := u.Connectors.Bindings.BoundAccounts(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("accounts bound to %s: %w", instance, err)
	}
	goLive, err := u.Connectors.GoLiveDates(ctx, bound)
	if err != nil {
		return nil, err
	}
	var out []FileAccount
	for _, a := range bound {
		only, err := boundInstance(ctx, u.Bindings, a)
		if errors.Is(err, ErrNoConnector) || only != instance {
			continue
		}
		if err != nil {
			return nil, err
		}
		_, seeded := goLive[a]
		out = append(out, FileAccount{Account: a, Instance: instance, Mode: ingest.UploadMode(seeded, 0)})
	}
	return out, nil
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

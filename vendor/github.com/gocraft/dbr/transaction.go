package dbr

import (
	"context"
	"database/sql"
	"time"
)

// Tx is a transaction for the given Session
type Tx struct {
	EventReceiver
	Dialect Dialect
	*sql.Tx
	Timeout time.Duration

	// normally we don't call the context cancelFunc.
	// however, if we start a tx without explictly tx,
	// we will need to call this after the transaction.
	Cancel func()
}

func (tx *Tx) GetTimeout() time.Duration {
	return tx.Timeout
}

func (sess *Session) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := sess.Connection.BeginTx(ctx, opts)
	if err != nil {
		return nil, sess.EventErr("dbr.begin.error", err)
	}
	sess.Event("dbr.begin")

	stx := &Tx{
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		Tx:            tx,
		Cancel:        func() {},
	}
	deadline, ok := ctx.Deadline()
	if ok {
		stx.Timeout = deadline.Sub(time.Now())
	}
	return stx, nil
}

// Begin creates a transaction for the given session
func (sess *Session) Begin() (*Tx, error) {
	ctx := context.Background()
	var cancel func()
	timeout := sess.GetTimeout()
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	stx, err := sess.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	if cancel != nil {
		stx.Cancel = cancel
	}
	return stx, nil
}

// Commit finishes the transaction
func (tx *Tx) Commit() error {
	defer tx.Cancel()
	err := tx.Tx.Commit()
	if err != nil {
		return tx.EventErr("dbr.commit.error", err)
	}
	tx.Event("dbr.commit")
	return nil
}

// Rollback cancels the transaction
func (tx *Tx) Rollback() error {
	defer tx.Cancel()
	err := tx.Tx.Rollback()
	if err != nil {
		return tx.EventErr("dbr.rollback", err)
	}
	tx.Event("dbr.rollback")
	return nil
}

// RollbackUnlessCommitted rollsback the transaction unless it has already been committed or rolled back.
// Useful to defer tx.RollbackUnlessCommitted() -- so you don't have to handle N failure cases
// Keep in mind the only way to detect an error on the rollback is via the event log.
func (tx *Tx) RollbackUnlessCommitted() {
	defer tx.Cancel()
	err := tx.Tx.Rollback()
	if err == sql.ErrTxDone {
		// ok
	} else if err != nil {
		tx.EventErr("dbr.rollback_unless_committed", err)
	} else {
		tx.Event("dbr.rollback")
	}
}

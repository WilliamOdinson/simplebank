package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is an interface to expose all functions from SQL queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// SQLStore provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	connPool *pgxpool.Pool
}

// NewStore creates a new Store
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}

// execTx executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, txFunc func(*Queries) error) error {
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	err = txFunc(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %w, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

// TransferTx performs a money transfer from one account to another.
// It creates a transfer record, add account entries, and update accounts' balance within a single db transaction.
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = execChangeBalance(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)

			if err != nil {
				return err
			}
		} else {
			result.ToAccount, result.FromAccount, err = execChangeBalance(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)

			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func execChangeBalance(ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64) (account1 Account, account2 Account, err error) {

	account1, err = q.ChangeAccountBalance(ctx, ChangeAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})

	if err != nil {
		return
	}

	account2, err = q.ChangeAccountBalance(ctx, ChangeAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})

	if err != nil {
		return
	}
	return
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

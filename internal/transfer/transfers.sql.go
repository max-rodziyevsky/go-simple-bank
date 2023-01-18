// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: transfers.sql

package transfer

import (
	"context"
)

const createTransfer = `-- name: CreateTransfer :one
insert into transfers(from_account_id, to_account_id, amount)
values ($1, $2, $3)
returning id, from_account_id, to_account_id, amount, created_at
`

type CreateTransferParams struct {
	FromAccountID int64 `db:"from_account_id" json:"from_account_id"`
	ToAccountID   int64 `db:"to_account_id" json:"to_account_id"`
	Amount        int64 `db:"amount" json:"amount"`
}

func (q *Queries) CreateTransfer(ctx context.Context, arg CreateTransferParams) (Transfer, error) {
	row := q.db.QueryRowContext(ctx, createTransfer, arg.FromAccountID, arg.ToAccountID, arg.Amount)
	var i Transfer
	err := row.Scan(
		&i.ID,
		&i.FromAccountID,
		&i.ToAccountID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const getTransfer = `-- name: GetTransfer :one
select id, from_account_id, to_account_id, amount, created_at
from transfers
where id = $1
limit 1
`

func (q *Queries) GetTransfer(ctx context.Context, id int64) (Transfer, error) {
	row := q.db.QueryRowContext(ctx, getTransfer, id)
	var i Transfer
	err := row.Scan(
		&i.ID,
		&i.FromAccountID,
		&i.ToAccountID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const listTransfers = `-- name: ListTransfers :many
select id, from_account_id, to_account_id, amount, created_at from transfers
where from_account_id = $1 OR to_account_id = $2
limit $3
offset $4
`

type ListTransfersParams struct {
	FromAccountID int64 `db:"from_account_id" json:"from_account_id"`
	ToAccountID   int64 `db:"to_account_id" json:"to_account_id"`
	Limit         int32 `db:"limit" json:"limit"`
	Offset        int32 `db:"offset" json:"offset"`
}

func (q *Queries) ListTransfers(ctx context.Context, arg ListTransfersParams) ([]Transfer, error) {
	rows, err := q.db.QueryContext(ctx, listTransfers,
		arg.FromAccountID,
		arg.ToAccountID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Transfer
	for rows.Next() {
		var i Transfer
		if err := rows.Scan(
			&i.ID,
			&i.FromAccountID,
			&i.ToAccountID,
			&i.Amount,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

-- name: CreateAccount :one
INSERT INTO accounts
(
    owner, balance, currency
)
VALUES
(
    $1, $2, $3
)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: ListAccounts :many
select * from accounts
order by id
limit $1
offset $2;

-- name: UpdateAccount :one
update accounts
set balance = $2
where id = $1
returning *;

-- name: DeleteAccount :exec
delete from accounts
where id = $1;
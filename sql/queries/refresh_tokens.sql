-- name: CreateRefreshToken :one
insert into refresh_tokens (
  id, created_at, updated_at, user_id, expires_at, revoked_at
) values (
  $1, now(), now(), $2, $3, NULL
)
returning *;

-- name: GetUserFromRefreshToken :one
select * from refresh_tokens
where id = $1;

-- name: RevokeRefreshTokenWithToken :exec
update refresh_tokens
set
  updated_at = now(),
  revoked_at = now()
where id = $1;

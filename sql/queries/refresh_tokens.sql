-- name: CreateRefreshToken :one
insert into refresh_tokens (
  id, created_at, updated_at, user_id, expires_at
) values (
  gen_random_uuid(), now(), now(), $1, $2
)
returning *;

-- name: GetUserFromRefreshToken :one
select * from refresh_tokens
where id = $1;

-- name: RevokeRefreshTokenWithToken :exec
update refresh_tokens
set
  updated_at = now(),
  expires_at = now()
where id = $1;

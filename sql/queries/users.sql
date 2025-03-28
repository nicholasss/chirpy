-- name: CreateUser :one
insert into users (
	id, created_at, updated_at, email, hashed_password, is_chirpy_red
) values (
	gen_random_uuid(), NOW(), NOW(), $1, $2, false
)
returning *;

-- name: UpdateUser :one
update users
set
  updated_at = now(),
  email = $2,
  hashed_password = $3
where id = $1
returning id, created_at, updated_at, email, is_chirpy_red;

-- name: ResetUsers :exec
delete from users;

-- name: GetUserByEmailRetHashedPassword :one
select * from users
where email = $1;

-- name: GetUserByEmailSafe :one
select id, created_at, updated_at, email, is_chirpy_red from users
where email = $1;

-- name: GetUserByIDSafe :one
select id, created_at, updated_at, email, is_chirpy_red from users
where id = $1;

-- name: UpgradeUserByID :exec
update users
set
  updated_at = now(),
  is_chirpy_red = true  
where id = $1;

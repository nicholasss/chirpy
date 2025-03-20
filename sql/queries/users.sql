-- name: CreateUser :one
insert into users (
	id, created_at, updated_at, email, hashed_password
) values (
	gen_random_uuid(), NOW(), NOW(), $1, $2
)
returning *;

-- name: UpdateUser :one
update users
set
  updated_at = now(),
  email = $2,
  hashed_password = $2
where id = $1
returning id, created_at, updated_at, email;

-- name: ResetUsers :exec
delete from users;

-- name: GetUserByEmailRetHashedPassword :one
select * from users
where email = $1;

-- name: GetUserByEmailSafe :one
select id, created_at, updated_at, email from users
where email = $1;

-- name: GetUserByIDSafe :one
select id, created_at, updated_at, email from users
where id = $1;

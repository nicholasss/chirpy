-- name: CreateUser :one
insert into users (
	id, created_at, updated_at, email, hashed_password
) values (
	gen_random_uuid(), NOW(), NOW(), $1, $2
)
returning *;

-- name: ResetUsers :exec
delete from users;

-- name: GetUserByEmailWHashedPassword :one
select * from users
where email = $1;

-- name: GetUserByEmailWOPassword :one
select id, created_at, updated_at, email from users
where email = $1;

-- name: GetUserByID :one
select id, created_at, updated_at, email from users
where id = $1;

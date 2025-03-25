# API Documentation

## Generic endpoints

- "GET /api/heathz"
  - Request:
    Nothing required
  - Response:
    If the server is online, Status 200.

## User endpoints

- "POST /api/users"
  Utilized for creating a user.

  - Request:

  ```json
  {
    "email": "<string: user email>",
    "password": "<string: raw password>"
  }
  ```

  - Response:

  ```json
  {
    "id": "<string: user uuid>",
    "created_at": "<string: timestamp>",
    "updated_at": "<string: timestamp>",
    "email": "<string: user email>",
    "is_chirpy_red": "<boolean>",
    "access_token": "", // blank
    "refresh_token": "" // blank
  }
  ```

- "PUT /api/users"
  Utilized to update a users password or email.

  - Request:
    Requires access token (JWT) in authorization header.

  ```json
  {
    "email": "<string: email>",
    "password": "<string: raw password>"
  }
  ```

  - Response:
    If the password was updated, it will not show.
    If there is a 2XX response then it was successful.

  ```json
  {
    "id": "<string: user uuid>",
    "created_at": "<string: timestamp>",
    "updated_at": "<string: timestamp>",
    "email": "<string: user email>",
    "is_chirpy_red": "<boolean>"
  }
  ```

- "POST /api/login"

  - Request:

  ```json
  {
    "email": "<string: email>",
    "password": "<string: raw password>"
  }
  ```

  - Response:

  ```json
  {
    "id": "<string: user uuid>",
    "created_at": "<string: timestamp>",
    "updated_at": "<string: timestamp>",
    "email": "<string: user email>",
    "is_chirpy_red": "<boolean>",
    "access_token": "<string: JWT/access token>",
    "refresh_token": "<string: refresh_token>"
  }
  ```

## Chirp endpoints

- "DELETE /api/chirps/{id}"
  - Request:
  - Response:
- "GET /api/chirps"
  - Request:
  - Response:
- "GET /api/chirps/{id}"
  - Request:
  - Response:
- "POST /api/chirps"
  - Request:
  - Response:

## Admin endpoints

- "GET /admin/metrics"
  - Request:
  - Response:

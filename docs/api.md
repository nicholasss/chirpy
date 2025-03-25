# API Documentation

## Generic endpoints

- "GET /api/heathz"
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
    If the password was updated, do not expect it in the response. If there the updated_at field is within the last ~5 seconds, it was updated

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
    No access token (JWT) is required.

  ```json
  {
    "email": "<string: email>",
    "password": "<string: raw password>"
  }
  ```

  - Response:
    Utilized to acquire a refresh token (good for 60 days), or an access token (JWT).

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

- "POST /api/refresh"

  - Request:
    Requires a valid refresh token in the authorization header.

  - Response:
    A new access token is provided, and the previous access token is revoked.
    TODO: Revoke previous access token on this endpoint.

  ```json
  {
    "access_token": "<string: JWT/access token>"
  }
  ```

- "POST /api/revoke"

  - Request:
    Requires a valid refresh token in the authorization header.

  - Response:
    If successful, a status 204 is expected. In order to get a new access token, you must use "POST /api/refresh" with a valid refresh toke with a valid refresh token.

- "POST /api/polka/webhooks"

  - Request:
    Requires a valid "polka" API key (polka is an analogue for a payment processor) in the authorization header. The server reads the environmental variable for "POLKA_KEY".

  ```json
  {
    "event": "user.upgrade",
    "data": {
      "user_id"
    }
  }
  ```

## Chirp endpoints

- "DELETE /api/chirps/{id}"
  Utilized to delete a specific chirp, as long as you are the author.

  - Request:
    Requires access token (JWT) in authorization header. Change '{id}' to be a specific chirp id.

  - Response:
    Expect a status 204 if successful.

- "GET /api/chirps"
  Utilized to request either all chirps, or chirps by a specific author.

  - Request:
    No access token (JWT) is required. In order to request all chirps by a specific author, use a query parameter.
    `/api/chirps?author_id=<author's user id>`

  - Response:
    Expect a list of chirp objects. If there was no `author_id` query parameter, then you will receive all chirps.

    ```json
    [
      {
        "id": "<string: chirp id>",
        "created_at": "<string: timestamp>",
        "updated_at": "<string: timestamp>",
        "body": "<string: body of chirp>",
        "user_id": "<string: authors user id>"
      },
      {
        ...
      }
    ]
    ```

- "GET /api/chirps/{id}"
  Utilized to request a specific chirp.

  - Request:
    No access token (JWT) is required. Change '{id}' to be a specific chirp id.

  - Response:

    ```json
    {
      "id": "<string: chirp id>",
      "created_at": "<string: timestamp>",
      "updated_at": "<string: timestamp>",
      "body": "<string: body of chirp>",
      "user_id": "<string: authors user id>"
    }
    ```

- "POST /api/chirps"
  Utilized for posting chirps from your user.

  - Request:
    Requires access token (JWT) in authorization header.

    ```json
    {
      "body": "<string>"
    }
    ```

- Response:
  Expect a status 201 if successful.

```json
{
  "id": "<string: chirp id>",
  "created_at": "<string: timestamp>",
  "updated_at": "<string: timestamp>",
  "body": "<string: body of chirp>",
  "user_id": "<string: authors user id>"
}
```

## Admin endpoints

- "POST /admin/reset"
  ! This endpoint is only available when the environmental variable "PLATFORM" is set to development. (It can be set to 'development' or 'production')
  Utilized for deleting all records from the database.

  - Request:
    No API keys are required.

  - Response:
    Expect a status 200 if successful. Nothing important is expected in the response body.

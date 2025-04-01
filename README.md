# Chirpy

_Chirpy is a guided project from boot.dev._

HTTP server project, written in Golang. The goal was to not only get the basics of an HTTP server, but to to also understand additional topics:

- Query Parameters
- Refresh/Access (JWT)
- Database Queries
- Middleware Request Processing
- Modern Password Storage Techniques
- Basic Webhook Implementation

## .env

There are a few items required to be read from the environment for this project to work.

- PLATFORM: `development` | `production`
- GOOSE_DRIVER: `postgres` | `<sql_db_type>`
- GOOSE_DBSTRING: URL of the database to connect to
- JWT_SECRET: Securely generated string used for signing JWT's

## API Documentation

The full documentation can be found [here](/docs/api.md) for the API.

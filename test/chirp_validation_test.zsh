#!/bin/zsh
# This should test the validation endpoint for chirps
# /api/validate_chirp

url="localhost:8080/api/validate_chirp"
header="Content-Type: application/json"

curl -d '{"body": "This is a short chirp."}' -H $header $url

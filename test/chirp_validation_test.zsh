#!/bin/zsh
# This should test the validation endpoint for chirps
# /api/validate_chirp

url="localhost:8080/api/validate_chirp"
header="Content-Type: application/json"

expected1='{"valid":true}'
response1=$(curl -s -d '{"body": "This is a short chirp."}' -H $header $url)

expected2='{"error":"Chirp is too long"}'
response2=$(curl -s -d '{"body": "This is a reeeeeeeeeeeeeeeeeeeeeeeaaaaaaaaaaaaaaallllllllllllllly looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong chirp."}' -H $header $url)

expectedList=(expected1 expected2)
responseList=(response1 response2)

for ((i = 1; i <= ${#expectedList[@]}; i++)); do
	if [ ${#expectedList[i]} != ${#responseList[i]} ]; then
		print " ### failure"
		print "expected: ${#expectedList[i]}"
		print "response: ${#responseList[i]}"
	else;
		print " ### success"
	fi
done

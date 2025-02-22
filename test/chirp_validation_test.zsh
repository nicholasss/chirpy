#!/bin/zsh
# This should test the validation endpoint for chirps
# /api/validate_chirp

url="localhost:8080/api/validate_chirp"
header="Content-Type: application/json"

expected1='{"valid":true}'
response1=$(curl -s -d '{"body": "This is a short chirp."}' -H $header $url)

expected2='{"error":"Chirp is too long"}'
response2=$(curl -s -d '{"body": "This is a reeeeeeeeeeeeeeeeeeeeeeeaaaaaaaaaaaaaaallllllllllllllly looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong chirp."}' -H $header $url)

expected3='{"error":"Something went wrong"}'
response3=$(curl -s -d '{"body: "This is invalid JSON."}' -H $header $url)

expected4='{"valid":true}'
response4=$(curl -s -d '{"body": "This is a long but still valid chirp. %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%"}' -H $header $url)

expectedList=(expected1 expected2 expected3 expected4)
responseList=(response1 response2 expected3 expected4)

for ((i = 1; i <= ${#expectedList[@]}; i++)); do
	if [ "${#expectedList[i]}" != "${#responseList[i]}" ]; then
		print " ### failure $i"
		print "expected: ${#expectedList[i]}"
		print "response: ${#responseList[i]}"
	else;
		print " ### success $i"
	fi
done

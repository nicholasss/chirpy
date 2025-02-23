#!/usr/bin/zsh

# sources the local .env file

if [ -e "./.env" ]; then
	print "Found local '.env' file. Sourcing..."
	source ./.env
else
	print "No local '.env' file found."
	print "Are you in the right directory, or is it missing?"
fi

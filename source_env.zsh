#!/usr/bin/zsh
# source this script in order for it to work as intended

# sources the local .env file
if [ -e "./.env" ]; then
	print "Found local '.env' file. Sourcing..."
	source ./.env
else
	print "No local '.env' file found."
	print "Are you in the right directory, or is it missing?"
fi

# # activates the venv environment
# if [ -d "./venv" ] && [ -e "./venv/bin/activate" ]; then
# 	print "Found local './venv/' directory. Activating..."
# 	source ./venv/bin/activate
# else
# 	print "No local './venv/' directory found."
# 	print "Are you in the right directory, or is it missing?"
# fi

#!/bin/zsh

echo " ### Running all tests now."
echo ""

for file in ./test/*.zsh;
do
	echo -n $(date +%H:%M:%S)
	echo " ### Testing: $file"
	"$file"
done

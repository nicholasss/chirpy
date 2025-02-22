#!/usr/bin/zsh

print " ### Running all tests now."
print ""

cd ~/Developer/Bootdev_Projects/chirpy

for file in $(find ./test -name '*_test.zsh' -executable); do
	print -n $(date +%H:%M:%S)
	print " ### Testing: $file"
	"$file"
done

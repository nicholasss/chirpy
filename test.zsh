#!/usr/bin/zsh

print " ### Running all tests now."

# health check
upurl="localhost:8080/api/healthz"
uptest=$(curl -s $upurl)
if [ $uptest != "OK" ]; then
	print " ### Server is offline."
	exit 1
else
	print " ### Server is online."
fi
print ""

cd ~/Developer/Bootdev_Projects/chirpy

for file in $(find ./test -name '*_test.zsh' -executable); do
	print -n " ### "
	print -n $(date +%H:%M:%S)
	print " $file"
	"$file"
done

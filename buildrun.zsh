#! /bin/zsh

echo "building project..."
go build -o out

if [[ $? -ne 0 ]]; then
	echo "error building project."
	echo ""
	exit 1
fi

echo "running project..."
./out

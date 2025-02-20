#! /bin/zsh

echo "building project..."
go build -o out

echo "running project..."
./out

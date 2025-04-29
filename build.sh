#!/usr/bin/env bash

echo "installing deps..."
sh ./dep.sh
echo "running tests..."

go test -timeout 600s -v ./src/galcone/...

buildCmd=""
echo $buildCmd

if [[ "$1" == "windows" ]]
then
    buildCmd="GOOS=windows go build -o"
elif [[ "$1" == "linux" ]]
then
    buildCmd="CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o"
elif [[ "$1" == "macos" ]]
then
    buildCmd="GOOS=darwin go build -o"
else
    buildCmd="go build -o"
fi
echo $buildCmd

# Run the build command
cmd="$buildCmd galcon ./src"
eval ${cmd}

# Check if the build was successful
if [[ $? -eq 0 ]]; then
    echo "Build was successful!"
else
    echo "Build failed!"
    exit 1  # Exit the script with a non-zero status to indicate failure
fi

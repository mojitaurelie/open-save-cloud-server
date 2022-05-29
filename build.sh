#!/bin/bash

platforms=("windows/amd64" "windows/arm64" "darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64")

if [[ -d "./build" ]]
then
    rm -r ./build
fi

mkdir build
cd build

for platform in "${platforms[@]}"
do
    echo "* Compiling for $platform..."
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='osc-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name -a ../main.go
done
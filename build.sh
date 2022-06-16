#!/bin/bash

platforms=("windows/amd64" "linux/amd64" "linux/arm64" "linux/arm")

if [[ -d "./build" ]]
then
    rm -r ./build
fi

mkdir build

for platform in "${platforms[@]}"
do
    echo "* Compiling for $platform..."
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='./build/osc-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
        go generate
        env GOAMD64=v3 GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=1 go build -o $output_name -a
    else
        if [ $GOARCH = "arm" ]; then
            env GOARM=7 GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -o $output_name -a
        else
            if [ $GOARCH = "amd64" ]; then
                env GOAMD64=v3 GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -o $output_name -a
            else
                env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -o $output_name -a
            fi
        fi
    fi

done
#!/bin/bash
CGO_ENABLED=0 go build -ldflags "-s -w"  -o excel2image .

docker build -t excel2image:0.1 .

rm excel2image

docker-compose -f docker-compose.yaml up -d

docker image prune -f
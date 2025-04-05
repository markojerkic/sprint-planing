#!/bin/bash
watchexec \
  --clear \
  -r \
  -e go,css,js \
  --no-vcs-ignore \
  --filter '*.go' \
  --filter 'cmd/web/assets/**/*' \
  -d 1000 \
  -- go run cmd/api/main.go

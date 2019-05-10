#!/bin/bash

gox -ldflags="-s -w" -osarch="darwin/amd64"
upx goccp_darwin_amd64

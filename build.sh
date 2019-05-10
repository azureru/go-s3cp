#!/bin/bash

gox -ldflags="-s -w" -osarch="linux/amd64 darwin/amd64 linux/arm"
upx goccp_*

#!/bin/bash

gox -osarch="darwin/amd64"
upx goccp_darwin_amd64

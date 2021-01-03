#!/bin/bash

BUILD=`date '+%Y/%m/%d %H:%M:%S'` go build -ldflags "-X main.buildTime=$BUILD" .
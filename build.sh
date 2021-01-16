#!/bin/bash

go build -ldflags "-X main.buildTime=$(date +"%Y.%m.%d.%H%M%S")" .
cp ./raylar /home/sinan/bin/raylar
#!/bin/bash
cd bin
go build  -o watchdogGo ../src/
if [ ! -f ./config.json ]; then
    cp ../src/config/config.json ./
fi

if [ ! -d ./mail ]; then
    cp -r ../src/notification/mail ./
fi

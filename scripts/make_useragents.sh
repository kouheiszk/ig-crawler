#!/bin/sh

prefix="package ua\nvar (UserAgents = []string{\n"
list=$(curl https://raw.githubusercontent.com/nlf/browser-agents/master/browsers.json | jq 'flatten[] | flatten[]' | sed -e 's/$/,/' | grep WebKit)
suffix="\n})"

echo "$prefix$list$suffix" > pkg/ua/list.go
go fmt pkg/ua/list.go

#!/bin/bash
export GO_WANT_HELPER_PROCESS=1
go test -run=TestHelperProcess github.com/KrushnaVardhanReddy/substrate/engine/internal/mcp &> output.log &
pid=$!
sleep 1
cat output.log
kill $pid 2>/dev/null

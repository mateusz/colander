#!/bin/bash -ex

trap "exit 0" SIGINT SIGTERM

while true; do
	CONC=$(( ( RANDOM % 6 )  + 4 ))
	TIME=$(( ( RANDOM % 50 )  + 10 ))
	siege -A "crawler" http://localhost:8888/slow.php\?f\=1 -c $CONC -t ${TIME}s
done

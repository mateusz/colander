#!/bin/bash -ex

trap "exit 0" SIGINT SIGTERM

for CONC in {1,5,8,7,5,2,5,2,6,9,5,3,2,10,5,10,1,4,6,1}; do
	siege -A "crawler" http://localhost:8888/slow.php\?f\=1 -c $CONC -t 20s
done

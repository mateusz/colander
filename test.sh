#!/bin/bash

./test_pseudo.sh &
PID=$!

ab -c 1 -n 100 -g ab.tsv http://localhost:8888/slow.php\?f\=1

killall siege
kill $PID

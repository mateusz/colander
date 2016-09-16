#!/bin/bash

ulimit -n 10000

if [ -z "$1" ]; then
	echo "Please provide name for the test that will be used as the graph directory name."
	exit 2
fi

./test_pseudo.sh &
PID=$!

ab -c 1 -n 200 -g ab.tsv http://localhost:8888/slow.php\?f\=1

killall siege
kill $PID

mkdir -p "$1"
gnuplot < graphs/distribplot.txt 
mv distrib.jpg "$1/distrib.jpg"
gnuplot < graphs/seqplot.txt 
mv seq.jpg "$1/seq.jpg"

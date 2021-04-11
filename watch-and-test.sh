#!/bin/bash

# This script watch files updates and trigger "go test"
# If test failed, it will ring a bell.

# inotify-tools is required for this script.

# parameter 1 is the path to watch
# if parameter 1 is not specified, it runs from current dir

cd ${1:-.}

go_test ()
{
	echo
	date
       	go test
	if [ $? != "0" ]; then
		echo -e '\a' 
	fi
}

pwd 

go_test

while true
do
	inotifywait -qq -e close_write .
	sleep 1 # work around intermittent 'no file to test' issue
	go_test
done

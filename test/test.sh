#!/bin/sh
if test "$(NUMEN_PIPE=/tmp/numen-test-pipe numen --audio=test.wav test.phrases)" = \
'n
u
m
e
n'
then
	echo PASSED TEST
else
	echo FAILED TEST
	exit 1
fi

#!/bin/sh
if test "$(NUMEN_PIPE=/tmp/numen_pipe_test numen --audio=test.wav test.phrases 2>&1)" = \
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

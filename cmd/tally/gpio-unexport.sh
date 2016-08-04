#!/bin/sh
#
# unexport gpio to clear

for gpio in "$@"; do
    [ -e /sys/class/gpio/gpio$gpio ] && echo $gpio > /sys/class/gpio/unexport
done

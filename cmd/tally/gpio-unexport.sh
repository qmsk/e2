#!/bin/sh
#
# unexport gpio to clear

[ "$#" -gt 0 ] || exit 0

for gpio in "$@"; do
    [ -e /sys/class/gpio/gpio$gpio ] && echo $gpio > /sys/class/gpio/unexport
done

#!/bin/sh
#
# raspian uses udev to assign /sys/class/gpio group write permissions, but this is racy between gpio export and writing to .../direction:
#   http://raspberrypi.stackexchange.com/questions/23162/gpio-value-file-appears-with-wrong-permissions-momentarily
# export beforehand and wait for udev before running

for gpio in "$@"; do
    [ -e /sys/class/gpio/gpio$gpio ] || echo $gpio > /sys/class/gpio/export
done

echo "$0: Waiting for gpio devices to settle..." >&2

udevadm settle

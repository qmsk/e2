#!/bin/bash
#
# Trvial bash script to shutdown SPI-LED output.
#
# Used for e.g. systemd [Service] ExecStopPost=...

count=$1
DEVICE=${2:-/dev/spidev0.0}

(
    # start frame
    printf '\0\0\0\0'

    # led frames
    for i in $(seq $count); do
        printf '\xC0\x00\x00\x00'
    done
    
    # stop frame
    # XXX: different protocol variants!
    printf '\0\0\0\0\0\0\0\0'

) > $DEVICE


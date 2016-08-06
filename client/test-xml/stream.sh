set -ue

target=$1
stream=${2:-}

cat send-reset.xml | nc -q -1 -v $target 9876 | (
    if [ -n "$stream" ]; then
        tee -a $stream
    else
        cat
    fi
)

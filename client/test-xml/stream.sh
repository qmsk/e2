set -ue

target=$1
stream=${2:-stream-$(date +%Y%m%d-%H%M%S).xml}

cat send-reset.xml | nc -q -1 -v $target 9876 | tee -a $stream

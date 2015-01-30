# Depends
* libyaml-dev
* qmsk-dmx (qmsk.web)

# Install

    apt-get install python3 libyaml-dev
    virtualenv -p python3 opt
    ./opt/bin/pip3 install -r requirements.txt

# Config

    mkdir etc

    curl -v http://192.168.0.201/backup-download | tar -C etc -xzv

# Usage

    PYTHONPATH=../qmsk-dmx ./opt/bin/python3 ./qmsk-e2-web \
        --e2-host 192.168.0.201 \
        --e2-presets-xml etc/xml/ \
        --e2-presets-db var/e2.db \
        --e2-web-port 8081 \
        --e2-web-static static/ \
        -v --debug-module qmsk.net.e2.presets


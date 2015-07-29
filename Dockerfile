FROM debian:jessie

RUN apt-get update
RUN apt-get -y install \
    python3 python3-pip python3-virtualenv \
    virtualenv

# build-from-source deps
RUN apt-get -y install \
    git npm nodejs-legacy
RUN npm install -g bower

# data
VOLUME /srv/qmsk-e2/
RUN adduser --system --uid 999 qmsk-e2 --home /srv/qmsk-e2

# virtualenv
RUN virtualenv -p python3 --system-site-packages /opt/qmsk-e2
ADD requirements.txt /tmp/qmsk-e2/
RUN /opt/qmsk-e2/bin/pip install -r /tmp/qmsk-e2/requirements.txt

# install
ADD . /tmp/qmsk-e2
RUN /opt/qmsk-e2/bin/pip install /tmp/qmsk-e2/dist/qmsk-dmx-0.9.tar.gz # XXX
RUN cd /tmp/qmsk-e2 && /opt/qmsk-e2/bin/python3 setup.py install

# build share data
RUN cd /opt/qmsk-e2/share && bower --allow-root --config.directory=static/bower_components update

WORKDIR /srv/qmsk-e2/
USER qmsk-e2

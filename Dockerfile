FROM debian:jessie

RUN apt-get update
RUN apt-get -y install \
    python3 python3-pip python3-virtualenv \
    virtualenv

VOLUME /srv/qmsk-e2/
RUN adduser --system --uid 999 qmsk-e2 --home /srv/qmsk-e2

RUN virtualenv -p python3 --system-site-packages /opt/qmsk-e2

ADD requirements.txt /tmp/qmsk-e2/
RUN /opt/qmsk-e2/bin/pip install -r /tmp/qmsk-e2/requirements.txt

ADD . /tmp/qmsk-e2
RUN cd /tmp/qmsk-e2 && /opt/qmsk-e2/bin/python3 setup.py install

WORKDIR /srv/qmsk-e2/
USER qmsk-e2

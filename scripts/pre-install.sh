#!/bin/sh
getent group orcus_exporter >/dev/null || groupadd -r orcus_exporter
getent passwd orcus_exporter >/dev/null || \
    useradd -r -g orcus_exporter -s /sbin/nologin \
    -c "Orcus Prometheus Exporter" -d /home/orcus_exporter -m orcus_exporter
exit 0
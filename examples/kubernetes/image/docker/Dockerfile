FROM fedora:22
MAINTAINER Michael Bishop <michael@advisorconnect.co>

RUN dnf autoremove -y && dnf distro-sync -y && dnf upgrade -y && dnf clean all -y

RUN dnf -y install  \
 https://download.postgresql.org/pub/repos/yum/9.5/fedora/fedora-23-x86_64/pgdg-fedora95-9.5-3.noarch.rpm

RUN dnf check-update && dnf -y install postgresql-jdbc postgresql95 postgresql95-plperl python3-postgresql \
 postgresql95-plpython3 \
 postgresql95-plpython postgresql95-devel postgresql95-server postgresql95-contrib postgresql95-python \
 hostname pguri95.x86_64 pg_stat_kcache95 plv8_95 pg_qualstats95 pg_top95 python3-psycopg2 \
 table_version95 cyanaudit95 citus_95 prefix95  pgbson95 pgagent_95  && dnf clean all -y

RUN useradd -ms /bin/bash stolon

EXPOSE 5431 5432 6431

ENTRYPOINT ["/usr/local/bin/run.sh"]

ADD run.sh bin/stolon-keeper bin/stolon-sentinel bin/stolon-proxy /usr/local/bin/

RUN chmod +x /usr/local/bin/stolon-keeper /usr/local/bin/stolon-sentinel /usr/local/bin/stolon-proxy /usr/local/bin/run.sh


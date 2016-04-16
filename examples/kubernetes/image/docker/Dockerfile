FROM fedora:22

RUN dnf check-update && dnf upgrade -y && dnf clean all -y

RUN dnf -y install  https://download.postgresql.org/pub/repos/yum/9.5/fedora/fedora-23-x86_64/pgdg-fedora95-9.5-3.noarch
.rpm

RUN dnf -y install postgresql-server-9.5 postgresql-9.5 hostname

RUN dnf clean all

RUN useradd -ms /bin/bash stolon

EXPOSE 5431 5432 6431

ENTRYPOINT ["/usr/local/bin/run.sh"]

ADD run.sh bin/stolon-keeper bin/stolon-sentinel bin/stolon-proxy /usr/local/bin/

RUN chmod +x /usr/local/bin/stolon-keeper /usr/local/bin/stolon-sentinel /usr/local/bin/stolon-proxy /usr/local/bin/run.sh


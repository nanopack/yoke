FROM nanobox/base

# install gcc and build tools
RUN echo deb http://apt.postgresql.org/pub/repos/apt/ trusty-pgdg main >> /etc/apt/sources.list.d/pgdg.list && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    apt-get update && \
    apt-get install postgresql-9.4


ADD ./yoke /usr/bin
ADD ./yokeadm/yokeadm /usr/bin
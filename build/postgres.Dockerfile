FROM postgres

### Root section #######################################################################################################
ARG USER
ARG USER_PASSWORD
RUN useradd -ms /bin/bash $USER

COPY build/postgres_init.sh /docker-entrypoint-initdb.d/init.sh

RUN sed -i '4i user='$USER /docker-entrypoint-initdb.d/init.sh
RUN sed -i '6i password='$USER_PASSWORD /docker-entrypoint-initdb.d/init.sh

### User section #######################################################################################################
USER postgres

ENV TZ="UTC"

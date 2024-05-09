##############
# MIGRATIONS #
##############
FROM alpine:3.15.10 as postgres-migrations
WORKDIR /usr/src/app

# Install wget and other dependencies, if necessary
RUN apk --no-cache add wget

# Get migrations from ConsenSys web3signer repository
RUN wget -q https://github.com/ConsenSys/web3signer/archive/refs/tags/24.2.0.tar.gz && \
    tar -xvf 24.2.0.tar.gz && \
    rm 24.2.0.tar.gz

############
# POSTGRES #
############
FROM postgres:14.1-bullseye


ENV POSTGRES_PASSWORD=password \
    PGPASSWORD=password \
    POSTGRES_USER=postgres \ 
    POSTGRES_DB=web3signer \
    INITDB_DIR=/docker-entrypoint-initdb.d/

RUN apt update && apt install -y rename

COPY --from=postgres-migrations /usr/src/app/web3signer-24.2.0/slashing-protection/src/main/resources/migrations/postgresql/* ${INITDB_DIR}

RUN rename 's/(\d+)(?=.*\.)/sprintf("%03d",$1)/eg' /docker-entrypoint-initdb.d/*

CMD ["postgres"]
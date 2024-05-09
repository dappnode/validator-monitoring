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

##########
# FLYWAY #
##########
FROM flyway/flyway:9.16.1-alpine

RUN apk update && apk --no-cache add postgresql-client
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh && rm -rf /flyway/sql/*
COPY --from=postgres-migrations /usr/src/app/web3signer-24.2.0/slashing-protection/src/main/resources/migrations/postgresql/* /flyway/sql/
ENTRYPOINT ["entrypoint.sh"]
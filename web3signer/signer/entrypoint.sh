#!/bin/bash

exec /opt/web3signer/bin/web3signer --http-listen-port=9000 \
    --http-listen-host=0.0.0.0 --http-host-allowlist="*" --http-cors-origins="*" \
    eth2 \
    --slashing-protection-db-url=jdbc:postgresql://postgres.web3signer.dappnode:5432/web3signer \
    --slashing-protection-db-username=postgres \
    --slashing-protection-db-password=password \
    --key-manager-api-enabled=true \
    --Xsigning-ext-enabled=true

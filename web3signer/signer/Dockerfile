FROM eclipse-temurin:17 as jre-build
WORKDIR /app
# TODO: update with official repo
RUN apt update && apt install git -y && git clone https://github.com/usmansaleem/web3signer.git
WORKDIR /app/web3signer 
RUN git fetch origin && git checkout sign_extension && ./gradlew clean assemble && tar -xzf ./build/distributions/web3signer-develop.tar.gz

##############
# WEB3SIGNER #
##############

FROM consensys/web3signer:24.2.0
# Copy web3signer keymanager binary
COPY --from=jre-build /app/web3signer/web3signer-develop /opt/web3signer
COPY entrypoint.sh /usr/bin/entrypoint.sh

EXPOSE 9000
ENTRYPOINT /bin/bash /usr/bin/entrypoint.sh
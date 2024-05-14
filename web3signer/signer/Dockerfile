FROM eclipse-temurin:17 as jre-build
WORKDIR /app
# TODO: update with official repo
RUN apt update && apt install git -y && git clone https://github.com/consensys/web3signer.git
WORKDIR /app/web3signer 
RUN git fetch origin && ./gradlew clean assemble && tar -xzf ./build/distributions/web3signer-develop.tar.gz
COPY entrypoint.sh /usr/bin/entrypoint.sh

EXPOSE 9000
ENTRYPOINT /bin/bash /usr/bin/entrypoint.sh
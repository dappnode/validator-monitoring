# validator-monitoring

## Description

This repository hosts the code for a validator monitoring system designed to receive, validate, and store signatures from various networks. It includes a JWT generator for easy token generation required for API access
to certain endpoints.

**Workflow in Dappnode:**
1. Dappnode's [Staking Brain](https://github.com/dappnode/StakingBrain) sends a `PROOF_OF_VALIDATION` signature request to the web3signer. [Details on the format and integration here](https://github.com/Consensys/web3signer/pull/982).
2. The Staking Brain wraps web3signer's response in a `SignatureRequest` object and sends it to the monitoring system defined in this repo.
3. The signature is then validated and stored in a MongoDB database by the monitoring system defined in this repo.

A `SignatureRequest` object has the following format:

```go
type SignatureRequest struct {
 Payload   string `json:"payload"`
 Pubkey    string `json:"pubkey"`
 Signature string `json:"signature"`
 Tag       Tag    `json:"tag"`
}
```

##  API

- `/signatures?network=<network>`:
  - `POST`: Sends an array of signatures to be validated and stored in the database. The request body must be a non empty array of "SignatureRequest" objects.
  - `GET`: Returns all signatures stored in the the database for which the user has access to. More on this on the [Authentication](#authentication) section.

### Authentication

The `GET /signatures` endpoint is protected by a JWT token, which must be included in the HTTPS request. This token should be passed in the Authorization header using the Bearer schema. The expected format is:

```text
Bearer <JWT token>
```

#### JWT requirements

To access the `GET /signatures` endpoint, the JWT must meet the following criteria:

- **Key ID** (`kid`): The JWT must include a kid claim in the header. The kid must be whitelisted in the monitoring system, and will be used to identify the pubkey used to verify the JWT signature.

As a nice to have, the JWT can also include the following claims as part of the payload:

- **Expiration time** (`exp`): The expiration time of the token, in Unix time. If no `exp` is provided, the token will be valid indefinitely.
- **Subject** (`sub`): Additional information about the user or entity behind the token. (e.g. an email address)

#### Generating the JWT

To generate a JWT token, you can use the `jwt-generator` tool included in this repository. The tool requires an RSA private key in PEM format to sign the token.
A keypair in PEM format can be generated using OpenSSL:

```sh
    openssl genrsa -out private.pem 2048
    openssl rsa -in private.pem -pubout -out public.pem
```

Once you have the private key, you can generate a JWT token using the `jwt-generator` tool:

```sh
    ./jwt-generator --private-key=path/to/private.pem --kid=your_kid_here --exp=24h --output=path/to/output.jwt
```

Note: Contact the dappnode team to whitelist your JWT "kid" and public key.

##  Validation Process

The process of validating the request and the signature follows the next steps:

1. Get network query parameter from the request: it is mandatory and must be one of "mainnet", "holesy", "gnosis", "lukso".
2. Decode and validate the request. The request body must be an array of SignatureRequest objects. Each object must have the following format:

```go
type SignatureRequest struct {
 Payload   string `json:"payload"`
 Pubkey    string `json:"pubkey"`
 Signature string `json:"signature"`
 Tag       Tag    `json:"tag"`
}
```

The payload must be encoded in base64 and must have the following format:

```go
type DecodedPayload struct {
 Type      string `json:"type"`
 Platform  string `json:"platform"`
 Timestamp string `json:"timestamp"`
}
```

3. The validators must be in status "active_on_going" according to a standard beacon node API, see <https://ethereum.github.io/beacon-APIs/#/Beacon/postStateValidators>:
   3.1 The signatures from the validators that are not in this status will be discarded.
   3.2 If in the moment of querying the beacon node to get the validator status the beacon node is down the signature will be accepted storing the validator status as "unknown" for later validation.
4. Only the signatures that have passed the previous steps will be validated. The validation of the signature will be done using the pubkey from the request.
5. Only valid signatures will be stored in the database.

##  Crons

There are 2 cron to ensure the system is working properly:

- `removeOldSignatures`: this cron will remove from the db signatures older than 30 days
- `updateSignaturesStatus`:
  - This cron will update the status of the validators that are in status "unknown" to "active_on_going" if the validator is active in the beacon node.
  - If the beacon node is down the status will remain as "unknown".
  - If the validator is not active the signature will be removed from the database.

## Database

The database is a mongo db that stores the signatures as BSON's. There are considered as unique the combination of the following fields: `network`, `pubkey`, `tag`. In order to keep the size of the database as small as possible there is a `entries` collection that stores the payload signature and decodedPayload of each request.

The BSON of each unique validator has the following format:

```go
bson.M{
    "pubkey":  req.Pubkey,
    "tag":     req.Tag,
    "network": network,
    "entries": bson.M{
            "payload":   req.Payload,
            "signature": req.Signature,
            "decodedPayload": bson.M{
                "type":      req.DecodedPayload.Type,
                "platform":  req.DecodedPayload.Platform,
                "timestamp": req.DecodedPayload.Timestamp,
            },
        },
 }
```

**Mongo db UI**

There is a express mongo db UI that can be accessed at `http://localhost:8080`. If its running in dev mode and the compose dev was deployed on a dappnode environment then it can be access through <http://ui.dappnode:8080>

## Environment variables

See `.env.example` file for the list of environment variables that can be set.

```env
MONGO_INITDB_ROOT_USERNAME=
MONGO_INITDB_ROOT_PASSWORD=
MONGO_DB_API_PORT=
API_PORT=
LOG_LEVEL=
MAX_ENTRIES_PER_BSON= # It is recommended to set a low value like 100 for this variable since mongo db has a limit of 16MB per document
BEACON_NODE_URL_MAINNET=
BEACON_NODE_URL_HOLESKY=
BEACON_NODE_URL_GNOSIS=
BEACON_NODE_URL_LUKSO=
```

## Development environment

To run the development environment with all the pieces of the system (web3signer, staking brain and listener with the required infra), then you can run it with the following command:

```bash
docker compose -f docker-compose.dev.yml up -d --scale brain=5
```

The flag `--scale brain=5` is optional and it will run 5 instances of the staking brain in order to simulate a real environment.

## Running the system

**Requirements:**

- docker
- docker-compose

**Steps:**

1. Clone the repository
2. Run `docker-compose up` (production) or `docker compose -f docker-compose.dev.yml up` (development) in the root directory of the repository
3. Access database UI at `http://localhost:8081`.

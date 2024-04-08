# validator-monitoring

## Description

This repository contains the code for the validator monitoring system. The system is designed to listen signatures request from different networks validate them and store the results in a database.

## Running the system

**Requirements:**

- docker
- docker-compose

**Steps:**

1. Clone the repository
2. Run `docker-compose up` (production) or `docker compose -f docker-compose.dev.yml up` (development) in the root directory of the repository
3. Access database UI at `http://localhost:8081`.

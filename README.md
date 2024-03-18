# trustdoc

Trustdoc is microservice which provides APIs for its clients to upload, verify and download files.

### Purpose:

The purpose of this microservice is to provide a way to store documents in a secure and tamper-proof way. It uses
blockchain to store the hash of the document which can be used to verify the integrity of the document. It also stores
the owner information of the document which can be used to verify the authenticity of the document.

### Features:

1. Upload a document.
2. Download a document.
3. Verify the integrity of the document.
4. Verify the authenticity of the document.(TODO - currently failing with an error)

It uses following infrastructure:

1. Postgres for storing document metadata.
2. Minio for storing documents as blob objects.
3. Kaleido for storing document hash in blockchain.

### Design:

1. All the components are developed in a pluggable fashion which means any implementation can be replaced if it
   implements
   the interface it exposes.
2. The microservice is developed in a way that it can be deployed in a distributed fashion. It can be scaled
   horizontally.

## Technical notes:

1. The microservice is developed in golang using go-gin.
2. It uses gorm as an ORM for postgres.
3. It uses minio-go for interacting with minio.
4. It uses low level functions like eth_sendRawTransaction to interact with Kaleido when installing a new contract.
5. It uses abigen to generate go bindings for the contract.
6. It supports http tracing and logging using structured json format.
7. It exposes swagger for its APIs.
8. It uses makefile for building and running the project along with its dependencies.
9. It uses docker for packaging the project.
10. It uses docker-compose for running the project along with its dependencies.

## Local step:-

1. Clone this repository.
2. cd into the directory
3. `make startAll` to run all along with its dependencies.
4. `make runApp` if you want to run the application only.
5. `docker-compose --env-file .env up` to run the application along with its dependencies.

Following 2 environment variables are required to run the application as it connects to Kaleido SaaS platform.

1. `KALEIDO_NODE_API_URL` - used to authenticate and talk to a node running on Kaleido which provides us the blockchain
   service.
2. `SIGN_PRIV_KEY` - used to sign the content before sending it to the blockchain.
 
## API Documentation:

## Integration test:
   Use newman to run the integration tests, the collection is present in the `docs` directory.

## TODO

    `make docker`
    to create a docker image for this project and then can be run through docker as well!

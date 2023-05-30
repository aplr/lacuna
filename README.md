# Lacuna - a Cloud Pub/Sub Docker Operator

Lacuna is a Kubernetes-style operator that runs locally on your machine and manages Google Cloud Pub/Sub topics and subscriptions for your docker containers. It is designed to run alongside a local [Pub/Sub emulator](https://cloud.google.com/pubsub/docs/emulator) and manages topics and subscriptions automatically by observing the docker containers running on your machine.

## Overview

While testing locally, it is often useful to have a local Pub/Sub emulator running to simulate the behavior of a real Pub/Sub instance. However, it can be tedious to manually create topics and subscriptions for each docker container that needs to interact with Pub/Sub. For Pub/Sub this is especially true because topics and subscriptions can not be created using the gcloud CLI, but must be created using a proper Pub/Sub API client. Lacuna aims to solve this problem by creating topics and subscriptions for each container that needs to interact with Pub/Sub just by using docker labels.

### Limitations

Currently, Lacuna only supports Google Cloud Pub/Sub, but it can be extended to support other messaging systems.

## Usage

```yaml
version: "3.9"

services:
    json-server:
        image: ghcr.io/clue/json-server:latest
        restart: unless-stopped
        labels:
            lacuna.enabled: true
            lacuna.subscription.test.topic: test
            lacuna.subscription.test.endpoint: http://json-server/messages
        ports:
            - "8080:8080"
    pubsub:
        image: gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators
        command: gcloud beta emulators pubsub start --host-port=0.0.0.0:8085 --project=pubsub
        restart: unless-stopped
        ports:
            - "8085:8085"
        volumes:
            - ./data/db.json:/data/db.json
    lacuna:
        image: ghcr.io/aplr/lacuna:latest
        restart: unless-stopped
        volumes:
            - /var/run/docker.sock:/var/run/docker.sock
```

## Configuration

Lacuna is configured using docker labels. The following labels are supported:

| Label                                 | Description                                                                        | Required |
| ------------------------------------- | ---------------------------------------------------------------------------------- | -------- |
| `lacuna.enabled`                      | Enables Lacuna for the container.                                                  | Yes      |
| `lacuna.subscription.<name>.topic`    | The name of the topic to subscribe to.                                             | Yes      |
| `lacuna.subscription.<name>.endpoint` | The endpoint to send messages to.                                                  | Yes      |
| `lacuna.subscription.<name>.deadline` | The number of seconds to wait for an acknowledgement before resending the message. | No       |

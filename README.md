# Puppy - a Cloud Pub/Sub Docker Operator

Puppy is a Google Cloud Pub/Sub operator for your local docker test environments, built with Go. It is designed to run alongside a local [Pub/Sub emulator](https://cloud.google.com/pubsub/docs/emulator) and manages topics and subscriptions automatically by observing the docker containers running on your machine.
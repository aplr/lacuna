FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators

ENV PUBSUB_PROJECT_ID=fruitsco

CMD [ "sh", "-c", "gcloud beta emulators pubsub start --project=${PUBSUB_PROJECT_ID}" ]
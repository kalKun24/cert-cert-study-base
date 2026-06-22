FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine

RUN apk add --no-cache openjdk21-jre-headless && \
    gcloud components install cloud-firestore-emulator --quiet

EXPOSE 8080

CMD ["gcloud", "emulators", "firestore", "start", "--host-port=0.0.0.0:8080", "--project=local-project"]

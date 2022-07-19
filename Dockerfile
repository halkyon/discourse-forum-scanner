FROM alpine:3.16.1@sha256:7580ece7963bfa863801466c0a488f11c86f85d9988051a9f9c68cb27f6b7872
COPY discourse-scanner /usr/bin/discourse-scanner
ENTRYPOINT ["/usr/bin/discourse-scanner"]

FROM alpine:3.15.4@sha256:4edbd2beb5f78b1014028f4fbb99f3237d9561100b6881aabbf5acce2c4f9454
COPY discourse-scanner /usr/bin/discourse-scanner
ENTRYPOINT ["/usr/bin/discourse-scanner"]

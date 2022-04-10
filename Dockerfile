FROM alpine
COPY discourse-scanner /usr/bin/discourse-scanner
ENTRYPOINT ["/usr/bin/discourse-scanner"]

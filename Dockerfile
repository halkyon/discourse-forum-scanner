FROM alpine:3.15.4@sha256:a777c9c66ba177ccfea23f2a216ff6721e78a662cd17019488c417135299cd89
COPY discourse-scanner /usr/bin/discourse-scanner
ENTRYPOINT ["/usr/bin/discourse-scanner"]

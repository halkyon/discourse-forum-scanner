FROM alpine
COPY discourse-forum-scanner /usr/bin/discourse-forum-scanner
ENTRYPOINT ["/usr/bin/discourse-forum-scanner"]

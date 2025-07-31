FROM alpine:latest

COPY privatecloud-cli /usr/local/bin/privatecloud-cli

ENTRYPOINT ["/usr/local/bin/privatecloud-cli", "pod"]

FROM scratch

COPY privatecloud-cli /

ENTRYPOINT ["/privatecloud-cli", "pod"]

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache git go npm yamllint
RUN git config --global --add safe.directory /app

RUN npm install -g @commitlint/cli @commitlint/config-conventional
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN npm install -g markdownlint-cli2 markdownlint-cli2-formatter-sarif

ENV PATH=$PATH:/root/go/bin

CMD ["./ci/lint.sh"]

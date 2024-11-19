FROM golang:alpine

WORKDIR /app

CMD ["./ci/test.sh"]

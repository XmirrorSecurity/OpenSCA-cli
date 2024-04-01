ARG VERSION
ARG TAG
ARG BASEIMAGE

FROM golang:latest AS builder

COPY ./ /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags "-s -w -X 'main.version=${VERSION}${TAG}' " \
    -o /app/opensca-cli main.go

FROM ${BASEIMAGE}
LABEL authors="OpenSCA Team <opensca@anpro-tech.com>"

COPY --from=builder /app/opensca-cli /app/opensca-cli
COPY --from=builder /app/opensca-cli /app/config.json
WORKDIR /src
ENTRYPOINT [ "/app/opensca-cli"]
CMD ["-path", "/src", "-config", "/src/config.json"]

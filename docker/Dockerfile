FROM golang:1.21.4 AS builder

ARG TARGETARCH
ARG VERSION

COPY ./ /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags "-X 'main.version=${VERSION}' " -o /app/opensca-cli main.go

FROM scratch
LABEL authors="OpenSCA Team <opensca@anpro-tech.com>"

COPY --from=builder /app/opensca-cli /app/opensca-cli
ENTRYPOINT [ "/app/opensca-cli"]
CMD ["-path", "/src", "-config", "/src/config.json"]

FROM alpine:latest

LABEL authors="OpenSCA Team <opensca@anpro-tech.com>"

ARG TARGETPLATFORM

VOLUME [ "/src" ]

WORKDIR /app

RUN apk add --update-cache curl jq && rm -rf /var/cache/apk/*

RUN if [ "${TARGETPLATFORM}" = "linux/amd64" ]; then \
    wget $(curl -s https://api.github.com/repos/XmirrorSecurity/OpenSCA-cli/releases/latest | \
    jq -r '.assets[] | select(.name | startswith("opensca-cli") and endswith("linux-amd64.tar.gz") ) | .browser_download_url') && \
    tar xvf *.tar.gz && rm ./*.tar.gz \
    ;fi

RUN if [ "${TARGETPLATFORM}" = "linux/arm64" ]; then \
    wget $(curl -s https://api.github.com/repos/XmirrorSecurity/OpenSCA-cli/releases/latest | \
    jq -r '.assets[] | select(.name | startswith("opensca-cli") and endswith("linux-arm64.tar.gz") ) | .browser_download_url') && \
    tar xvf *.tar.gz && rm ./*.tar.gz \
    ;fi

ENTRYPOINT [ "/app/opensca-cli", "-path", "/src", "-config", "/src/config.json"]
CMD [""]

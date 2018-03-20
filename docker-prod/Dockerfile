FROM alpine:latest

ENV CT_PROVIDER_VERSION "v0.2.1"

RUN apk update && \
    apk add ca-certificates git bash libc6-compat

RUN set -x \
    && export PROVIDER="terraform-provider-ct-${CT_PROVIDER_VERSION}-linux-amd64" \
    && apk --update add --no-cache --virtual .curl-deps curl \
    && curl -L -o /tmp/${PROVIDER}.tar.gz \
               https://github.com/coreos/terraform-provider-ct/releases/download/${CT_PROVIDER_VERSION}/${PROVIDER}.tar.gz \
    && tar -C /tmp -xzf /tmp/${PROVIDER}.tar.gz \
    && cp /tmp/${PROVIDER}/terraform-provider-ct /usr/local/bin \
    && rm -rf /tmp/${PROVIDER}* \
    && apk del .curl-deps

COPY terraform/* /usr/local/bin/
COPY check in out /opt/resource/

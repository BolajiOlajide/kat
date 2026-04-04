FROM alpine:3.21

ARG TARGETPLATFORM

RUN apk add --no-cache ca-certificates && \
    addgroup -S kat && \
    adduser -S -G kat kat

COPY ${TARGETPLATFORM}/kat /usr/local/bin/kat

USER kat

ENTRYPOINT ["/usr/local/bin/kat"]

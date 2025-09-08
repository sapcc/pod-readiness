FROM golang:1.24-alpine3.22 as builder
WORKDIR /go/src/github.com/sapcc/pod-readiness
RUN apk add --no-cache make git
COPY . .
ARG VERSION
RUN make all

FROM alpine:3.22
LABEL maintainer="Stefan Hipfel <stefan.hipfel@sap.com>"
LABEL source_repository="https://github.com/sapcc/pod-readiness"

ENV PACKAGES="curl"

RUN apk --no-cache add \
    $PACKAGES \
    && \
    rm -f /usr/lib/*.a && \
    (rm "/tmp/"* 2>/dev/null || true) && \
    (rm -rf /var/cache/apk/* 2>/dev/null || true)

WORKDIR /
RUN curl -Lo /bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64 \
    && chmod +x /bin/dumb-init \
    && dumb-init -V
COPY --from=builder /go/src/github.com/sapcc/pod-readiness/bin/linux/pod_readiness /usr/local/bin/
ENTRYPOINT ["dumb-init", "--"]
CMD ["pod_readiness"]

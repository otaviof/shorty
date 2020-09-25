#
# Build
#

FROM golang:1.15-alpine AS builder

ENV GO_DOMAIN="github.com" \
    GO_GROUP="otaviof" \
    GO_PROJECT="shorty"

ENV APP_DIR="${GOPATH}/src/${GO_DOMAIN}/${GO_GROUP}/${GO_PROJECT}"

RUN apk --update add gcc git make musl-dev

RUN mkdir -v -p ${APP_DIR}
WORKDIR ${APP_DIR}

COPY . ./
RUN make vendor install

#
# Run
#

FROM alpine:latest

ENV GO_PROJECT="shorty"

ENV USER_UID="1111" \
    GIN_MODE="release" \
    SHORTY_DATA="/var/lib/shorty" \
    SHORTY_DATABASE_FILE="/var/lib/shorty/shorty.sqlite" \
    SHORTY_ADDRESS="0.0.0.0:8000"

RUN apk --update add bash
COPY --from=builder /go/bin/${GO_PROJECT} /usr/local/bin/${GO_PROJECT}

RUN adduser -h ${SHORTY_DATA} -D -u ${USER_UID} ${GO_PROJECT}
VOLUME ${SHORTY_DATA}
WORKDIR ${SHORTY_DATA}}

ENTRYPOINT [ "/usr/local/bin/shorty" ]

USER ${USER_UID}

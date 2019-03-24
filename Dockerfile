#
# Build
#

FROM golang:1.12-alpine AS builder

ENV GO_DOMAIN="github.com" \
    GO_GROUP="otaviof" \
    GO_PROJECT="shorty"

ENV APP_DIR="${GOPATH}/src/${GO_DOMAIN}/${GO_GROUP}/${GO_PROJECT}"

RUN apk --update add gcc git make musl-dev
RUN go get -u github.com/golang/dep/cmd/dep

RUN mkdir -v -p ${APP_DIR}
WORKDIR ${APP_DIR}

COPY Makefile Gopkg.* ./
RUN make clean clean-vendor bootstrap

COPY . ./
RUN make

#
# Run
#

FROM golang:1.12-alpine

ENV GO_DOMAIN="github.com" \
    GO_GROUP="otaviof" \
    GO_PROJECT="shorty"

ENV APP_DIR="${GOPATH}/src/${GO_DOMAIN}/${GO_GROUP}/${GO_PROJECT}" \
    SHORTY_DATA="/var/lib/shorty"

RUN apk --update add bash
COPY --from=builder ${APP_DIR}/build/${GO_PROJECT} /usr/local/bin/${GO_PROJECT}

RUN mkdir -v -p ${SHORTY_DATA}}
VOLUME ${SHORTY_DATA}

WORKDIR /
ENTRYPOINT [ "/usr/local/bin/shorty" ]

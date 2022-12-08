ARG GO_VERSION_TAG="1.19-alpine"
FROM golang:${GO_VERSION_TAG} as build

WORKDIR /go/src/github.com/imup-io/client
COPY . .

RUN go mod download

ARG HONEYBADGER_API_KEY
ARG IMUP_CLIENT_VERSION
ARG NDT7_CLIENT_NAME
ARG BUILD_LD_FLAGS=-ldflags="-X 'main.ClientVersion=$(IMUP_CLIENT_VERSION)' -X 'main.ClientName=$(NDT7_CLIENT_NAME)' -X 'main.HoneybadgerAPIKey=$(HONEYBADGER_API_KEY)'"

RUN go build go build $(BUILD_LD_FLAGS) -o go/bin/client

FROM gcr.io/distroless/static-debian11 as imup

COPY --from=build go/bin/client /

CMD ["/imup"]

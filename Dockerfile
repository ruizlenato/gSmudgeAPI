##################### Build executable binary
FROM golang:alpine AS builder

WORKDIR /go/src/app
COPY . .

RUN apk update && apk add --no-cache git
RUN go get -d -v
RUN go build -o /go/bin/gSmudgeAPI


##################### Build Alpine Image
FROM alpine:latest

COPY --from=builder /go/bin/gSmudgeAPI /go/bin/gSmudgeAPI

RUN apk --no-cache add ca-certificates

EXPOSE 6969
ENTRYPOINT ["/go/bin/gSmudgeAPI"]
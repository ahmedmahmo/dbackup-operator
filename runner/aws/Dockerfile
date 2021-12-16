FROM golang:1.16 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY utils utils
COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o aws-runner main.go

FROM alpine:3.6 as alpine

RUN apk add -U --no-cache ca-certificates


FROM postgres
WORKDIR /
COPY --from=builder /workspace/aws-runner .
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER 65532:65532

ENTRYPOINT ["/aws-runner"]
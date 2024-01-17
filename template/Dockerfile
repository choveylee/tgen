FROM golang:1.21.5 as builder

ENV GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./{{app_name2}} ./cmd

FROM alpine:3.19.0

WORKDIR /app

RUN apk update \
    && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

COPY --chown=0:0 --from=builder /app/{{app_name2}} /app/{{app_name2}}
COPY --chown=0:0 --from=builder /app/cmd/{{app_name2}}_config.ini /app/{{app_name2}}_config.ini
COPY --chown=0:0 --from=builder /app/migration /app/migration

CMD ["/app/{{app_name2}}"]
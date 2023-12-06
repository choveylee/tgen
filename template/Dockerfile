FROM golang

ENV GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . .

RUN cp -R ./migration ./
RUN cp cmd/{{app_name2}}_config.ini ./
RUN go build -o ./{{app_name2}} ./cmd

CMD ["/app/{{app_name2}}"]
FROM golang:1.17 AS builder
#ENV GOSUMDB=off
#ENV GO111MODULE=on
#ENV GOPATH ""
#ENV GOPROXY=https://goproxy.cn,direct
RUN go env
WORKDIR /go/src

# compile
COPY . .

RUN GOPROXY=https://goproxy.cn go mod download
#RUN rm go.sum
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o . main.go
RUN make prod-build


# copy binary file
FROM alpine:3.9.4
RUN mkdir /app \
     # 日志路径，务必和应用中日志路径保持一致
    && mkdir /var/log/httpserver \
    && addgroup -g 10001 httpserver \
    && adduser -S -u 10001 -G httpserver httpserver
COPY --from=builder /go/src/app /app/app
COPY --from=builder /go/src/cert /app/cert/

RUN chown -R 10001:10001 /app \
    && chown -R 10001:10001 /var/log/httpserver

USER httpserver
WORKDIR /app
ENTRYPOINT ["./app"]

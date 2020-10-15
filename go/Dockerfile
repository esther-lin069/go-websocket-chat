FROM golang:1.15.2-alpine3.12

RUN go env -w GO111MODULE="on"
RUN apk add git
RUN go get -u github.com/gin-gonic/gin@v1.6.3 \
    && go get github.com/gorilla/websocket@v1.4.2 \
    && go get github.com/go-sql-driver/mysql@v1.5.0 \
    && go get github.com/go-redis/redis/v8@v8.3.0 

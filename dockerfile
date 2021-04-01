FROM golang:1.12-alpine as builder
RUN apk --no-cache add ca-certificates git
WORKDIR /build/myapp

# Fetch dependencies
COPY ./go.mod ./
RUN go mod download

# Build
COPY . ./
RUN CGO_ENABLED=0 go build ./apiWebServer/apiServer.go
RUN CGO_ENABLED=0 go build ./htmlWebServer/webserver.go

# Create final image
FROM alpine as apiServer
WORKDIR /root
COPY --from=builder /build/myapp/apiServer .
EXPOSE 8081
CMD ["./apiServer"]


FROM alpine as webserver
WORKDIR /root
COPY --from=builder /build/myapp/webserver .
COPY --from=builder /build/myapp/htmlWebServer/html/mainPage.html ./html/mainPage.html
COPY --from=builder /build/myapp/htmlWebServer/html/userPage.html ./html/userPage.html
EXPOSE 8080
CMD ["./webserver"]

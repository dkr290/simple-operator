FROM golang:alpine

ARG VERSION
ENV VERSION=$VERSION

WORKDIR /app
COPY main.go .
RUN go build -o app main.go

EXPOSE 8080
CMD ["./app"]

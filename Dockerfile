FROM golang:1.20-alpine as builder
LABEL authors="mohammadhoseinaref"

WORKDIR /app
COPY . .
RUN go build -o server api/server/main.go
RUN ls

FROM alpine:3.17
COPY --from=builder /app/server /server
RUN ls
CMD ./server
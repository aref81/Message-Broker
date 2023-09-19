FROM golang:latest as builder
LABEL authors="mohammadhoseinaref"


RUN apt -y update


RUN apt install -y protobuf-compiler


RUN apt install -y make
RUN apt install -y build-essential

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod tidy
COPY . .
CMD go run ./api/server/main.go
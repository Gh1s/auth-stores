FROM golang:1.15-alpine AS build

ARG STORE
ENV GO11MODULE on
ENV CGO_ENABLED 1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /src
ADD . .
WORKDIR /src/grpc/$STORE
RUN go mod download
RUN go build -ldflags="-s -w" -o /app/grpc.exe

FROM alpine:3.12 as final
ARG STORE
WORKDIR /app
COPY --from=build /app /app
COPY --from=build /src/grpc/$STORE/config.yaml /app/config.yaml
ENTRYPOINT ["./grpc.exe"]
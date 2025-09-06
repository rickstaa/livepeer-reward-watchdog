FROM golang:1.25.1-alpine AS build
WORKDIR /app
COPY . .
RUN cd scripts && go run download-abis.go
RUN go build -o reward-watcher main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/reward-watcher .
COPY --from=build /app/ABIs ./ABIs
ENTRYPOINT ["/app/reward-watcher"]

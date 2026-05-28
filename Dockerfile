# 可选：内网 CI / 无 Node 环境运行 kuaimai-cli
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN ./scripts/fetch_meta/fetch_meta.sh && \
    CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" -o /kuaimai-cli .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /kuaimai-cli /usr/local/bin/kuaimai-cli
ENTRYPOINT ["kuaimai-cli"]
CMD ["--help"]

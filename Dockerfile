FROM golang:1.19-alpine3.16 AS build
WORKDIR /build
COPY . /build
ARG COMMIT_HASH
ENV COMMIT_HASH=${COMMIT_HASH}
RUN apk add --no-cache gcc musl-dev && \
	go build -o mineshspc.com -ldflags "-s -w -linkmode external -extldflags -static -X main.Commit=$COMMIT_HASH -X 'main.BuildTime=`date '+%b %_d %Y, %H:%M:%S'`'"

FROM alpine:3.16
RUN apk add --no-cache ca-certificates
COPY --from=build /build/mineshspc.com /mineshspc.com

EXPOSE 8090
WORKDIR /data
ENTRYPOINT ["/mineshspc.com"]

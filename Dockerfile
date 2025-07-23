FROM --platform=$BUILDPLATFORM golang:1.20 as builder

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o unix-user-exporter .

FROM --platform=$TARGETPLATFORM alpine:latest
RUN apk --no-cache add procps
WORKDIR /root/
COPY --from=builder /app/unix-user-exporter .
EXPOSE 32142
ENTRYPOINT ["./unix-user-exporter"]

#force rebuild docker build --no-cache -t tdewin/stock-demo:latest .
FROM golang AS compiler
ENV CGO_ENABLED=0
RUN go install github.com/tdewin/stock-demo@latest && chmod 755 /go/bin/stock-demo

FROM alpine
LABEL maintainer="@tdewin"
WORKDIR /usr/sbin/
COPY --from=compiler /go/bin/stock-demo /usr/sbin/stock-demo
EXPOSE 8080
CMD /usr/sbin/stock-demo
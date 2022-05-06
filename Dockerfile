#force rebuild docker build --no-cache -t tdewin/stock-demo:latest .
FROM golang AS compiler
ENV CGO_ENABLED=0
#ENV GOPROXY=direct
#Causes other stuff not to work
ENV NOCACHECLONE=/go/src/github.com/tdewin
RUN mkdir -p $NOCACHECLONE && cd $NOCACHECLONE && git clone https://github.com/tdewin/stock-demo.git && cd stock-demo && go install . && chmod 755 /go/bin/stock-demo
#RUN go install github.com/tdewin/stock-demo@latest && chmod 755 /go/bin/stock-demo


FROM alpine
LABEL maintainer="@tdewin"
WORKDIR /usr/sbin/
COPY --from=compiler /go/bin/stock-demo /usr/sbin/stock-demo
RUN mkdir -p /var/stockdb/
EXPOSE 8080
CMD /usr/sbin/stock-demo
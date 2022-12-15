FROM alpine:3.15.2

RUN apk add --no-cache tzdata

ENV TZ Asia/Shanghai

WORKDIR $GOPATH/src/output

COPY output .

CMD ./output
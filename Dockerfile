FROM registry.cn-hangzhou.aliyuncs.com/oliujunk/alpine:3.21.3

RUN apk add --no-cache tzdata

ENV TZ Asia/Shanghai

WORKDIR $GOPATH/src/output

COPY output .

CMD ./output
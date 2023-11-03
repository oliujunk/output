# output

## 编译

```shell
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

$env:GOOS="linux"
$env:GOARCH="amd64"
go build
```

## 构建镜像

```shell
sudo docker build -t registry.cn-hangzhou.aliyuncs.com/oliujunk/output .

sudo docker push registry.cn-hangzhou.aliyuncs.com/oliujunk/output




sudo docker build -t registry.cn-hangzhou.aliyuncs.com/oliujunk/output-wang .

sudo docker push registry.cn-hangzhou.aliyuncs.com/oliujunk/output-wang
```

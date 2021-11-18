## websocekt 请求

    ws://10.10.10.200:8887/test3                可以请求K线图的1min 5min(等) 数据
    ws://10.10.10.200:8887/getRedisData         可以请求K线图的1min,行情step1 数据
    ws://10.10.10.200:8887/quotation            可以请求K线图的1min,行情step1 数据  不安全

## http请求

    http://10.10.10.200:8887/start              访问后开始缓存K线图1min 行情step1  K线历史1min 数据
    http://10.10.10.200:8887/klineHistory       访问后开始缓存K线历史1min 数据
    http://10.10.10.200:8887/getKlineHistory    访问后得到对应参数的K线图历史数据 ex. http://10.10.10.200:8887/getKlineHistory?symbol=btcusdt&period=5min

## 打包

```shell
GOOS=linux GOARCH=amd64 go build -o huobiService main.go
```
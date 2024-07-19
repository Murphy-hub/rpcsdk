# RPC-SDK

## RPC-SDK

go语言版本的微服务调用的SDK

#### client 初始化

```go
// 获取对象，head是请求的HEAD字段，用来解析HEAD中的trace/token等信息
client, err := rpcsdk.GetNewClient(header)
// 如果自身是顶层引用则header参数传递nil
header := http.Header{}
client, err := rpcsdk.GetNewClient(header)

// 调用JsInternal - Example
// client可以只调用一次，不需要在每次请求的上下文中重新new
client, err := gosdk.GetNewClient(header)
// 准备请求数据
dataSlice := []string{"xxx-token"}
data, _ := json.Marshal(dataSlice)
var result Token
err := client.CallParse(&gosdk.RequestParameter{
    Method: "post",
    Api:    "http://jsinternal/token/getDetail",
	Data:   data
}, &result)

// 如果需要自行解析响应体
resp, err := client.Call(&gosdk.RequestParameter{
    Method: "post",
    Api:    "http://jsinternal/token/getDetail",
    Data:   data
})
```

#### client方法

```go
type Client interface {
    Call(param *RequestParameter) ([]byte, error)             // 请求下层服务, 返回全部响应体，不做响应解析
    CallParse(param *RequestParameter, res interface{}) error // 请求下层服务, 依据内部rpc响应格式解析
    SetHeader(header http.Header)                 // 设置请求Header
    GetServer() *Server                           // 获取自身服务信息
    Release()                                     // 释放连接池
}
```

#### server 使用方法

```go
server := client.GetServer()

// 根据Header信息获取，如没有设置则获取不到
server.GetTraceId()     // 获取调用链TraceId
server.GetHeader()      // 获取Header
server.GetAppId()       // 获取自身appid
server.GetFromAppId()   // 获取请求来源appid
server.GetAccountId()   // 获取请求用户ID
```

#### 各个请求方法的参数结构体

```go
/**
 * Call、Request发起请求
 */
type RequestParameter struct {
	Method       string                 `json:"method"`
	Url          string                 `json:"api"`
	Data         []byte                 `json:"data"`
	ContentType  string                 `json:"content_type"`
	Header       map[string]string      `json:"header"`  // 需要写入请求信息的header数据
}
```


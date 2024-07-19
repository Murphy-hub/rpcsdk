package rpcsdk

import (
	"github.com/Murphy-hub/rpcsdk/errors"
	"net/http"
	"sync"
)

type Client interface {
	Call(param *RequestParameter) ([]byte, error)             // 请求下层服务, 返回全部响应体，不做响应解析
	CallParse(param *RequestParameter, res interface{}) error // 请求下层服务, 依据内部rpc响应格式解析
	SetHeader(header http.Header)
	GetServer() *Server
	Release()
}

var err error
var clientPool sync.Pool

func init() {
	clientPool.New = func() interface{} {
		return NewClient()
	}
}

func GetNewClient(header http.Header) (Client, error) {
	if header == nil {
		return nil, errors.BadRequest("The request is not valid")
	}
	cli := clientPool.Get().(Client)
	cli.SetHeader(header)
	return cli, err
}

func Release(cli Client) {
	cli.Release()
	clientPool.Put(cli)
}

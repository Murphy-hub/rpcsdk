package rpcsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Murphy-hub/rpcsdk/errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// REQUEST_HEADER 需要透传的请求头
var (
	REQUEST_HEADER = []string{
		"sw8",
		"accept-language",
		"x-real-ip",
		"X-Request-Id",
		"Childorgid",
		"access-token",
	}
)

type AppInfo struct {
	appId   string
	version string
}

// ResultCMO 标准业务响应格式解析
// {"code":0, "message":"ok", "obj":{}}
// {"code":0, "msg":"ok", "obj":{}}
// {"code":0, "err":"ok", "obj":{}}
type ResultCMO struct {
	Code    int32            `json:"code"`
	Message string           `json:"message"`
	Msg     string           `json:"msg"`
	Err     string           `json:"err"`
	Obj     *json.RawMessage `json:"obj"`
}

// ResultCEO 兼容旧业务，响应解析
// {"code": 0, "msg":"success", "obj": []}
type ResultCEO struct {
	C int32              `json:"c"`
	E string             `json:"e"`
	O []*json.RawMessage `json:"o"`
}

// CeoError C-E-O模式，业务错误
// {"code":0, "message":"ok", "obj":{}}
// {"code":0, "msg":"ok", "obj":{}}
// {"code":0, "err":"ok", "obj":{}}
type CeoError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Msg     string `json:"msg"`
	Err     string `json:"err"`
}

type rpcClient struct {
	header      http.Header
	currentInfo AppInfo
	targetInfo  AppInfo
	server      *Server
}

func NewRpcClient(header http.Header) (Client, error) {
	client := &rpcClient{
		header: header,
	}
	return client, err
}

func NewClient() Client {
	client := &rpcClient{}
	return client
}

func (c *rpcClient) checkParam(method, contentType string) error {
	method = strings.ToLower(method)
	if !InArray(method, AllowMethods) {
		return errors.MethodNotAllowed("method not allowed")
	}
	if !InArray(contentType, []string{ContentTypeForm, ContentTypeJson}) {
		return errors.BadRequest("content type error")
	}
	return nil
}

var _httpClient *http.Client

func (c *rpcClient) getHTTPClient() *http.Client {
	if _httpClient == nil {
		_httpClient = &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second, // 连接超时时间
					KeepAlive: 60 * time.Second, // 保持长连接的时间
				}).DialContext, // 设置连接的参数
				MaxIdleConns:          500,              // 最大空闲连接
				IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
				ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
				MaxIdleConnsPerHost:   100,              // 每个host保持的空闲连接数
			},
		}
	}
	return _httpClient
}

func (c *rpcClient) parseBody(
	method,
	reqUrl string,
	data []byte,
	contentType string,
) (req *http.Request, err error) {
	method = strings.ToUpper(method)

	switch contentType {
	case ContentTypeForm:
		theData := url.Values{}
		dataMap := make(map[string]interface{})
		err = json.Unmarshal(data, &dataMap)
		if err != nil {
			err = errors.BadRequest(err.Error())
			return
		}
		for k, v := range dataMap {
			theData.Set(k, fmt.Sprint(v))
		}
		body := strings.NewReader(theData.Encode())
		req, err = http.NewRequest(method, reqUrl, body)
		if err != nil {
			err = errors.BadRequest(err.Error())
			return
		}
		req.Header.Set("Content-Type", ContentTypeForm)
	case ContentTypeJson:
		body := bytes.NewReader(data)
		req, err = http.NewRequest(method, reqUrl, body)
		if err != nil {
			err = errors.BadRequest(err.Error())
			return req, err
		}
		req.Header.Set("Content-Type", ContentTypeJson)
	}
	return
}

func (c *rpcClient) Exec(
	method,
	reqURL string,
	data []byte,
	contentType string,
	header map[string]string,
) (out []byte, resp *http.Response, err error) {
	req, err := c.parseBody(method, reqURL, data, contentType)
	if err != nil {
		return
	}
	return c.exec(req, header)
}

func (c *rpcClient) exec(
	req *http.Request,
	header map[string]string,
) (out []byte, resp *http.Response, err error) {
	// 请求的header
	for k, v := range header {
		req.Header.Set(k, v)
	}
	// header透传
	if c.header != nil {
		for _, key := range REQUEST_HEADER {
			if val := c.header.Get(key); val != "" {
				req.Header.Set(key, val)
			}
		}
	}

	req.Header.Set("User-Agent", UserAgent+"/"+VERSION)
	req.Header.Set("Accept", "application/json")

	traceid := c.server.GetHeader(RequestTraceKey)
	if traceid != "" {
		req.Header.Set(RequestTraceKey, traceid)
		req.Header.Set(YxtTraceKey, traceid)
	}

	resp, err = c.getHTTPClient().Do(req)
	if err != nil {
		err = errors.InternalServerError(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		out, err = io.ReadAll(resp.Body)
		if err != nil {
			err = errors.InternalServerError(err.Error())
		}
	} else {
		err = requestError(resp)
	}
	return
}

func (c *rpcClient) Call(param *RequestParameter) (out []byte, err error) {
	if param.Method == "" {
		param.Method = AllowMethods[0]
	}
	if param.ContentType == "" {
		param.ContentType = ContentTypeJson
	}

	err = c.checkParam(param.Method, param.ContentType)
	if err != nil {
		return
	}
	out, _, err = c.Exec(param.Method, param.Url, param.Data, param.ContentType, param.Header)
	return
}

func (c *rpcClient) CallParse(param *RequestParameter, res interface{}) error {
	out, err := c.Call(param)
	if err != nil {
		return err
	}
	//out := []byte(`{"code": 0, "msg":"success", "obj": []}`)
	preParse := make(map[string]*json.RawMessage)
	err = json.Unmarshal(out, &preParse)
	if err != nil {
		return errors.InternalServerError("Response body Pre Unmarshal error: " + err.Error()).WithBody(out)
	}
	_, okc := preParse["c"]
	_, oko := preParse["o"]
	if okc && oko {
		return c.ParseCeo(out, res)
	}
	return c.ParseCmo(out, res)
}

// ParseCeo 解析 C-E-O模式响应
func (c *rpcClient) ParseCeo(out []byte, res interface{}) error {
	var resCeo ResultCEO
	err = json.Unmarshal(out, &resCeo)
	if err != nil {
		return errors.InternalServerError("Response body Unmarshal error: " + err.Error()).WithBody(out)
	}

	if resCeo.C == 0 {
		if res == nil {
			return nil
		}
		if len(resCeo.O) != 2 {
			return errors.InternalServerError("Remote service response information incomplete, expected length 2").WithBody(out)
		}
		// 解析响应数组第一个对象 - 业务异常信息
		if resCeo.O[0] != nil {
			var ceoError CeoError
			err = json.Unmarshal(*resCeo.O[0], &ceoError)
			if err != nil {
				return errors.InternalServerError("Business exception information parsing failed: " + err.Error()).WithBody(out)
			}
			return errors.InternalServerError(ceoError.Msg + ceoError.Message + ceoError.Err).WithBody(out)
		}

		// 解析业务数据
		if resCeo.O[1] != nil {
			err = json.Unmarshal(*resCeo.O[1], &res)
			if err != nil {
				return errors.InternalServerError("Business data parsing failed: " + err.Error()).WithBody(out)
			}
			return nil
		}

		// 无业务响应信息
		return nil
	} else {
		// rpc调用层响应异常
		return errors.InternalServerError(resCeo.E).WithBody(out)
	}
}

// ParseCmo 解析 Code-Msg-Obj 模式响应
func (c *rpcClient) ParseCmo(out []byte, res interface{}) error {
	var resCmo ResultCMO
	err := json.Unmarshal(out, &resCmo)
	if err != nil {
		return errors.InternalServerError("Response body Unmarshal error: " + err.Error()).WithBody(out)
	}
	if resCmo.Code == 0 {
		// 解析业务数据
		err = json.Unmarshal(*resCmo.Obj, &res)
		if err != nil {
			return errors.InternalServerError("Business data parsing failed: " + err.Error()).WithBody(out)
		}
		return nil
	} else {
		// rpc调用层响应异常
		return errors.InternalServerError(resCmo.Msg + resCmo.Message + resCmo.Err).WithBody(out)
	}
}

func (c *rpcClient) SetHeader(header http.Header) {
	c.header = header
	c.server = NewServer(header)
}

func (c *rpcClient) GetServer() *Server {
	return c.server
}

func (c *rpcClient) Release() {
	c.header = nil
	c.currentInfo = AppInfo{}
	c.targetInfo = AppInfo{}
	c.server = nil
}

// Call 新建一个 client 并请求解析响应
func Call(ctx context.Context, params *RequestParameter, response interface{}) error {
	cli, err := GetNewClient(http.Header{})
	if err != nil {
		return err
	}
	defer cli.Release()
	res, err := cli.Call(params)
	if err != nil {
		return err
	}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return err
	}
	return nil
}

// CallParse 新建一个client 请求并解析obj对象
func CallParse(ctx context.Context, params *RequestParameter, response interface{}) error {
	cli, err := GetNewClient(http.Header{})
	if err != nil {
		return err
	}
	defer cli.Release()
	err = cli.CallParse(params, response)
	if err != nil {
		return err
	}
	return nil
}

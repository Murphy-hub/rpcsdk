package rpcsdk

type RequestParameter struct {
	Method      string            `json:"method"`
	Url         string            `json:"url"`
	Data        []byte            `json:"data"`
	ContentType string            `json:"content_type"`
	Header      map[string]string `json:"header"`
}

type ResponseResult struct {
	Code int    `json:"code"`
	Obj  string `json:"obj"`
}

func (c ResponseResult) Success() bool {
	return c.Code == 0
}

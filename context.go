package imola

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Req *http.Request
	// Resp 如果用户直接使用，那就绕开了RespData、RespStatusCode这两个，会导致部分middleware无法运作
	Resp http.ResponseWriter
	// RespData 主要是为了 middleware读写用的
	RespData       []byte
	RespStatusCode int
	PathParams     map[string]string

	queryValues  url.Values
	MatchedRoute string
}

// BindJSON 解决大多数人的需求即可
func (c *Context) BindJSON(val any) error {
	if val == nil {
		return errors.New("web：输入不能为nil")
	}
	if c.Req.Body == nil {
		return errors.New("web: req body不能为nil")
	}
	decoder := json.NewDecoder(c.Req.Body)
	return decoder.Decode(val)
}

// FromValue 获取表单中的字段值
func (c *Context) FromValue(key string) StringValue {
	err := c.Req.ParseForm()
	if err != nil {
		return StringValue{
			val: "",
			err: err,
		}
	}
	/*vals, ok := c.Req.Form[key]
	if !ok {
		return "", errors.New("web: key不存在")
	}
	return vals[0], nil*/
	// 原始的api
	return StringValue{
		val: c.Req.FormValue(key),
		err: err,
	}
}

// QueryValue 处理请求url中的参数，例子：http://localhost:8081/form?name=xiaoming&age=18
func (c *Context) QueryValue(key string) StringValue {
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}
	// 用户区别不出来：是真的有值、还是恰好有空的字符串 (缓存一下)
	vals, ok := c.queryValues[key]
	if !ok {
		return StringValue{
			val: "",
			err: errors.New("web: key不存在"),
		}
	}
	return StringValue{
		val: vals[0],
		err: nil,
	}
	//return c.queryValues.Get(key), nil
}

func (c *Context) PathValue(key string) StringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return StringValue{
			err: errors.New("web: key不存在"),
		}
	}
	return StringValue{
		val: val,
	}
}

type StringValue struct {
	val string
	err error
}

func (s StringValue) AsInt64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.val, 10, 64)
}

func (c *Context) RespJSON(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.Resp.WriteHeader(status)
	/*n, err := c.Resp.Write(data)
	if n != len(data) {
		return errors.New("web: 未写入全部数据")
	}
	return err
	*/
	c.RespData = data
	c.RespStatusCode = status
	return nil
}

func (c *Context) RespJSONOK(val any) error {
	return c.RespJSON(http.StatusOK, val)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp, cookie)
}

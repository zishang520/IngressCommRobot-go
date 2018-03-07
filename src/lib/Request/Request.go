package Request

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/zishang520/persistent-cookiejar"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Construct struct {
	Proxy         string
	Header        map[string]string
	CookieFile    string
	CookieStore   bool
	NoPersist     bool
	IgnoreDiscard bool
	IgnoreExpires bool
}

type Request struct {
	Proxy       string
	Header      map[string]string
	Cookie      *cookiejar.Jar
	CookieStore bool
}

type Options struct {
	Method string
	Url    string
	Query  map[string][]string
	Header map[string]string
	Body   io.Reader
}

type Response struct {
	*http.Response
}

func (this Response) GetBody() ([]byte, error) {
	// defer this.Body.Close()
	return ioutil.ReadAll(this.Body)
}

func (this Response) CloseBody() error {
	return this.Body.Close()
}

type Json map[string]interface{}

func New(this *Construct) (request *Request) {
	request = &Request{}
	request.Cookie, _ = cookiejar.New(&cookiejar.Options{Filename: this.CookieFile, NoPersist: this.NoPersist, IgnoreDiscard: this.IgnoreDiscard, IgnoreExpires: this.IgnoreExpires})
	request.CookieStore = this.CookieStore
	request.Proxy = this.Proxy
	request.Header = this.Header
	return request
}

// 二次封装的请求方法
func (this *Request) Request(options *Options) (res *Response, err error) {
	var (
		response  *http.Response // 请求头
		request   *http.Request  // 请求头
		noOptions Options
	)
	if options == nil {
		options = &noOptions
	}
	client := &http.Client{Jar: this.Cookie}
	if len(this.Proxy) > 0 {
		dialer, err := proxy.SOCKS5("tcp", this.Proxy, nil, proxy.Direct)
		if err != nil {
			return res, err
		}
		client.Transport = &http.Transport{Dial: dialer.Dial}
	}
	if len(options.Query) > 0 {
		if strings.Contains(options.Url, "?") {
			options.Url = options.Url + "&" + url.Values(options.Query).Encode()
		} else {
			options.Url = options.Url + "?" + url.Values(options.Query).Encode()
		}
	}
	request, err = http.NewRequest(strings.ToUpper(options.Method), options.Url, options.Body)
	if err != nil {
		return res, err
	}
	if len(this.Header) > 0 {
		for key, value := range this.Header {
			request.Header.Set(key, value)
		}
	}
	if len(options.Header) > 0 {
		for key, value := range options.Header {
			request.Header.Set(key, value)
		}
	}
	if _, ok := request.Header["Content-Type"]; strings.EqualFold(strings.ToUpper(options.Method), "POST") && options.Body != nil && !ok {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	}
	response, err = client.Do(request)
	if err != nil {
		return res, err
	}
	if this.CookieStore {
		if err = this.Cookie.Save(); err != nil {
			return res, err
		}
	}
	// 解压gzio
	if strings.EqualFold(response.Header.Get("Content-Encoding"), "gzip") {
		response.Body, err = gzip.NewReader(response.Body)
		if err != nil {
			return res, err
		}
	}
	return &Response{response}, err
}

/**
 * 拼接字符串
 */
func (this *Request) Join(v []string, splite string) string {
	var buf bytes.Buffer
	for _, v := range v {
		if buf.Len() > 0 {
			buf.WriteString(splite)
		}
		buf.WriteString(v)
	}
	return buf.String()
}

/**
 * json数据
 */
func (this *Request) Json_decode(body []byte) (_json Json, err error) {
	if err := json.Unmarshal(body, &_json); err != nil {
		return _json, err
	}
	return _json, nil
}

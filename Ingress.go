package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zishang520/persistent-cookiejar"
	"io"
	"io/ioutil"
	"lib/Config"
	"lib/Set"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	// "reflect"
	// "path/filepath"
	"errors"
	"flag"
	"golang.org/x/net/proxy"
	"regexp"
	"strings"
	"time"
)

type Ingress struct {
	Proxy   string
	Header  map[string]string
	Mintime int
	Config  *Config.Options
	Cookie  *cookiejar.Jar
	Sqlite3 *sql.DB
}

type Options struct {
	Method string
	Url    string
	Header map[string]string
	Body   io.Reader
}

type Response struct {
	Status           string // e.g. "200 OK"
	StatusCode       int    // e.g. 200
	Proto            string // e.g. "HTTP/1.0"
	ProtoMajor       int    // e.g. 1
	ProtoMinor       int    // e.g. 0
	Header           http.Header
	Body             []byte
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Uncompressed     bool
	Trailer          http.Header
	Request          *http.Request
}

type Json map[string]interface{}

const COOKIE_FILE = "./data/cookie.json"
const AGENT_DB = "./data/agent.db"
const CONF_FILE = "./data/conf.json"
const TMP_FILE = "./data/tmp.json"

func New(mintime int) (ingress *Ingress, err error) {
	var v string
	fmt.Println("Initialized...")
	ingress = &Ingress{}
	fmt.Println("Open Sqlite")
	ingress.Sqlite3, err = sql.Open("sqlite3", AGENT_DB)
	if err != nil {
		return ingress, err
	}
	fmt.Println("Initialized Cookiejar")
	ingress.Cookie, err = cookiejar.New(&cookiejar.Options{Filename: COOKIE_FILE, NoPersist: false, IgnoreDiscard: true, IgnoreExpires: true})
	if err != nil {
		return ingress, err
	}
	fmt.Println(fmt.Sprintf("%s %d\n", "Set msg time", mintime))
	ingress.Mintime = mintime
	fmt.Println("Get Config")
	ingress.Config, err = ingress.GetConf()
	if err != nil {
		return ingress, err
	}
	fmt.Println("Set Proxy")
	ingress.Proxy = ingress.Config.Get("proxy").(string)
	fmt.Println("Initialized Default Header")
	ingress.Header = map[string]string{
		"Cache-Control":             "no-cache, no-store",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		"Accept-Encoding":           "gzip",
		"Accept-Language":           "zh-CN,zh;q=0.8",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Dnt":        "1",
		"User-Agent": ingress.Config.Get("UA").(string),
	}
	fmt.Println("Set V")
	v, err = ingress.__get_user_v()
	if err != nil {
		return ingress, err
	}
	ingress.Config.Set("v", v)
	fmt.Println("Set Token")
	ingress.Header["X-CSRFToken"], err = ingress.__get_token()
	if err != nil {
		return ingress, err
	}
	fmt.Println("Initialized successfully")
	return ingress, err
}

func (I *Ingress) GetConf() (conf *Config.Options, err error) {
	conf, err = Config.New(CONF_FILE)
	if err != nil {
		return conf, errors.New("Load Config Error")
	}
	if !conf.Has("UA") {
		return conf, errors.New("undefined index UA or value is not string or value is empty")
	}
	if !conf.Has("email") {
		return conf, errors.New("undefined index email or value is not string or value is empty")
	}
	if !conf.Has("password") {
		return conf, errors.New("undefined index password or value is not string or value is empty")
	}
	if !conf.Has("minLatE6") {
		return conf, errors.New("undefined index minLatE6 or value is not int|string or value is empty")
	}
	if !conf.Has("minLngE6") {
		return conf, errors.New("undefined index minLngE6 or value is not int|string or value is empty")
	}
	if !conf.Has("maxLatE6") {
		return conf, errors.New("undefined index maxLatE6 or value is not int|string or value is empty")
	}
	if !conf.Has("maxLngE6") {
		return conf, errors.New("undefined index maxLngE6 or value is not int|string or value is empty")
	}
	if !conf.Has("latE6") {
		return conf, errors.New("undefined index latE6 or value is not int|string or value is empty")
	}
	if !conf.Has("lngE6") {
		return conf, errors.New("undefined index lngE6 or value is not int|string or value is empty")
	}
	return conf, nil
}

// 二次封装的请求方法
func (I *Ingress) Request(options *Options) (res Response, err error) {
	var (
		body      []byte         //相应内容
		response  *http.Response //相应内容
		request   *http.Request  // 请求头
		noOptions Options
	)
	if options == nil {
		options = &noOptions
	}
	client := &http.Client{Jar: I.Cookie}
	if I.Proxy != "" {
		dialer, err := proxy.SOCKS5("tcp", I.Proxy, nil, proxy.Direct)
		if err != nil {
			return res, err
		}
		client.Transport = &http.Transport{Dial: dialer.Dial}
	}
	request, err = http.NewRequest(strings.ToUpper(options.Method), options.Url, options.Body)
	if err != nil {
		return res, err
	}
	for key, value := range I.Header {
		request.Header.Set(key, value)
	}
	for key, value := range options.Header {
		request.Header.Set(key, value)
	}
	if _, ok := request.Header["Content-Type"]; options.Body != nil && !ok {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	}
	response, err = client.Do(request)
	if err != nil {
		return res, err
	}
	defer response.Body.Close()
	if err = I.Cookie.Save(); err != nil {
		return res, err
	}
	// 解压gzio
	if _, ok := response.Header["Content-Encoding"]; ok && response.Header.Get("Content-Encoding") == "gzip" {
		response.Body, err = gzip.NewReader(response.Body)
		if err != nil {
			return res, err
		}
	}
	body, err = ioutil.ReadAll(response.Body)
	return Response{
		Status:           response.Status,
		StatusCode:       response.StatusCode,
		Proto:            response.Proto,
		ProtoMajor:       response.ProtoMajor,
		ProtoMinor:       response.ProtoMinor,
		Header:           response.Header,
		Body:             body,
		ContentLength:    response.ContentLength,
		TransferEncoding: response.TransferEncoding,
		Close:            response.Close,
		Uncompressed:     response.Uncompressed,
		Trailer:          response.Trailer,
		Request:          response.Request,
	}, err
}

func (I *Ingress) __get_token() (s string, err error) {
	for _, v := range I.Cookie.AllCookies() {
		if reg := regexp.MustCompile("(?sim:^csrftoken=(\\w+))").FindStringSubmatch(v.String()); len(reg) == 2 {
			return reg[1], err
		}
	}
	return s, errors.New("Failed to get token")
}

// 获取v
func (I *Ingress) __get_user_v() (v string, err error) {
	conf, err := Config.New(TMP_FILE)
	if conf.Has("time") && I.__diff_date(int64(conf.Get("time").(float64))) >= 0 {
		return conf.Get("v").(string), err
	} else {
		v, err = I.__get_v()
		if err != nil {
			return v, err
		}
		if _, e := conf.Set("v", v).Set("time", time.Now().Unix()).Save(); e != nil {
			return v, e
		}
	}
	return v, err
}

func (I *Ingress) __get_v() (r string, err error) {

	response, err := I.Request(&Options{
		Method: "GET",
		Url:    "https://www.ingress.com/intel",
	})
	if err != nil {
		return r, err
	}
	if response.StatusCode != 200 {
		return r, errors.New("Request Error")
	}
	if reg := regexp.MustCompile("(?sim:<a\\shref=\"(?P<URL>.*?)\"\\s.*?>Sign\\sin</a>)").FindAllSubmatch(response.Body, -1); len(reg) > 0 && len(reg[0]) == 2 && len(reg[0][1]) > 0 {
		if !I.__login(string(reg[0][1])) {
			return r, errors.New("Auto Login error")
		}
		response, err = I.Request(&Options{
			Method: "GET",
			Url:    "https://www.ingress.com/intel",
		})
		if err != nil {
			return r, err
		}
		if response.StatusCode != 200 {
			return r, errors.New("Request Error")
		}
	}
	if v := regexp.MustCompile("(?sim:<script\\stype=\"text/javascript\"\\ssrc=\"/jsc/gen_dashboard_(\\w+)\\.js\"></script>)").FindAllSubmatch(response.Body, -1); len(v) > 0 && len(v[0]) == 2 && len(v[0][1]) > 0 {
		return string(v[0][1]), err
	}
	return r, errors.New("Failed to get V")
}

func (I *Ingress) __check_islogin(body []byte) bool {
	return !regexp.MustCompile("(?sim:登录|login)").Match(body)
}

func (I *Ingress) __chaeck_refresh(body []byte) (string, bool) {
	if reg := regexp.MustCompile("(?sim:<meta\\s+http-equiv=\"refresh\"\\s+content=\"\\d+;\\s+url=(.*?)\">)").FindAllSubmatch(body, -1); len(reg) == 2 && len(reg[0]) == 2 {
		return strings.Replace(string(reg[0][1]), "&amp;", "&", -1), true
	}
	return "", false
}

func (I *Ingress) __refresh(_url, _referer string) (Response, bool) {
	response, err := I.Request(&Options{
		Method: "GET",
		Url:    _url,
		Header: map[string]string{
			"Origin":  "https://accounts.google.com",
			"Referer": _referer,
		},
	})
	if err != nil || response.StatusCode != 200 {
		return response, false
	}
	if _u, ok := I.__chaeck_refresh(response.Body); ok {
		return I.__refresh(_u, _url)
	}
	return response, true
}

func (I *Ingress) __login(_url string) bool {
	fmt.Println("Auto Login...")
	response, err := I.Request(&Options{
		Method: "GET",
		Url:    _url,
	})
	if err != nil || response.StatusCode != 200 {
		return false
	}
	if I.__check_islogin(response.Body) {
		return true
	}
	if _u, ok := I.__chaeck_refresh(response.Body); ok {
		if response, ok = I.__refresh(_u, _url); !ok {
			return false
		}
	}
	if err != nil || response.StatusCode != 200 {
		return false
	}
	document, err := goquery.NewDocumentFromReader(bytes.NewBuffer(response.Body))
	if err != nil {
		return false
	}
	username_xhr_url := "https://accounts.google.com/signin/v1/lookup"
	CheckEmailData := make(url.Values)
	document.Find("form[action]").Each(func(i int, contentSelection *goquery.Selection) {
		if v, ok := contentSelection.Attr("action"); ok {
			username_xhr_url = v
			contentSelection.Find("input[name][value]").Each(func(i int, contentSelection *goquery.Selection) {
				name, name_ok := contentSelection.Attr("name")
				value, value_ok := contentSelection.Attr("value")
				if (name_ok && value_ok) && len(name) > 0 && len(value) > 0 {
					CheckEmailData.Set(name, value)
				}
			})
		}
	})
	CheckEmailData.Set("Email", I.Config.Get("email").(string))

	post_response, err := I.Request(&Options{
		Method: "POST",
		Url:    username_xhr_url,
		Header: map[string]string{
			"Origin":  "https://accounts.google.com",
			"Referer": response.Request.URL.String(),
		},
		Body: strings.NewReader(CheckEmailData.Encode()),
	})
	if err != nil || post_response.StatusCode != 200 {
		return false
	}

	document, err = goquery.NewDocumentFromReader(bytes.NewBuffer(post_response.Body))
	if err != nil {
		return false
	}
	password_url := "https://accounts.google.com/signin/challenge/sl/password"
	LoginData := make(url.Values)
	document.Find("form[action]").Each(func(i int, contentSelection *goquery.Selection) {
		if v, ok := contentSelection.Attr("action"); ok {
			password_url = v
			contentSelection.Find("input[name][value]").Each(func(i int, contentSelection *goquery.Selection) {
				name, name_ok := contentSelection.Attr("name")
				value, value_ok := contentSelection.Attr("value")
				if (name_ok && value_ok) && len(name) > 0 && len(value) > 0 {
					LoginData.Set(name, value)
				}
			})
		}
	})
	LoginData.Set("Email", I.Config.Get("email").(string))
	LoginData.Set("Passwd", I.Config.Get("password").(string))
	login_page_response, err := I.Request(&Options{
		Method: "POST",
		Url:    password_url,
		Header: map[string]string{
			"Origin":  "https://accounts.google.com",
			"Referer": post_response.Request.URL.String(),
		},
		Body: strings.NewReader(LoginData.Encode()),
	})
	if err != nil || login_page_response.StatusCode != 200 {
		return false
	}
	if I.__check_islogin(login_page_response.Body) {
		return true
	}
	return false
}

func (I *Ingress) __diff_date(date int64) int64 {
	time1 := int64(time.Now().Unix() / 100)
	time2 := int64(date / 100)
	return int64((time2 - time1) / 864)
}

func (I *Ingress) __check_new_agent(msg string) (string, bool) {
	var newAgent string
	if reg := regexp.MustCompile("\\[secure\\]\\s+(\\w+):\\s+has\\scompleted\\straining\\.").FindAllStringSubmatch(msg, -1); len(reg) == 1 && len(reg[0]) == 2 {
		newAgent = reg[0][1]
	} else if reg := regexp.MustCompile("(?sim:\\[secure\\]\\s(\\w+):\\s+.*)").FindAllStringSubmatch(msg, -1); len(reg) == 1 && len(reg[0]) == 2 {
		if regexp.MustCompile(I.__regexp()).MatchString(msg) {
			newAgent = reg[0][1]
		}
	}
	if newAgent != "" {
		rows, _ := I.Sqlite3.Query("SELECT COUNT(`id`) AS num FROM `user` WHERE `agent`=\"" + newAgent + "\" LIMIT 1")
		var num int
		for rows.Next() {
			_ = rows.Scan(&num)
		}
		if num > 0 {
			return "", false
		}
		return newAgent, true
	}
	return "", false
}

func (I *Ingress) __join(v []interface{}, splite string) string {
	var buf bytes.Buffer
	for _, v := range v {
		if buf.Len() > 0 {
			buf.WriteString(splite)
		}
		buf.WriteString(v.(string))
	}
	return buf.String()
}

func (I *Ingress) __regexp() string {
	if I.Config.Has("regexp") && len(I.Config.Get("regexp").([]interface{})) > 0 {
		return "(" + I.__join(I.Config.Get("regexp").([]interface{}), "|") + ")"
	}
	return "(大家好|我是萌新|新人求带|新人求罩|大佬们求带|求组织|带带我)"
}

func (I *Ingress) __rand_msg() string {
	data := []interface{}{
		" 欢迎新人，快来加入川渝蓝军群(群号126821831)，发现精彩内容。",
		" 欢迎选择加入抵抗军·川渝蓝军群(群号126821831)，一起为建设社会主义社会、实现人类的全面自由发展而奋斗吧。",
		" 您已进入秋名山路段，此处常有老司机出没，加入川渝蓝军群(群号126821831)，寻找这里的老司机吧。",
		" 欢迎加入熊猫抵抗军(群号126821831)，感谢你在与shapers的斗争中选择了人性与救赎，选择与死磕并肩同行。新人你好，我是死磕。",
		" ingrees亚洲 中国分区 川渝地区组织需要你！快来加入川渝蓝军群(群号126821831)。",
	}
	rand.Seed(time.Now().UnixNano())
	if I.Config.Has("rand_msg") && len(I.Config.Get("rand_msg").([]interface{})) > 0 {
		data = I.Config.Get("rand_msg").([]interface{})
	}
	return data[rand.Intn(len(data))].(string)
}

func (I *Ingress) get_msg() (_json Json, err error) {
	Data, err := json.Marshal(map[string]interface{}{
		"minLatE6":       I.Config.Get("minLatE6").(float64),
		"minLngE6":       I.Config.Get("minLngE6").(float64),
		"maxLatE6":       I.Config.Get("maxLatE6").(float64),
		"maxLngE6":       I.Config.Get("maxLngE6").(float64),
		"minTimestampMs": (time.Now().Unix()*1000 - 60000*int64(I.Mintime)),
		"maxTimestampMs": -1,
		"tab":            "faction",
		"ascendingTimestampOrder": true,
		"v": I.Config.Get("v").(string),
	})
	if err != nil {
		return _json, err
	}
	response, err := I.Request(&Options{
		Method: "POST",
		Url:    "https://www.ingress.com/r/getPlexts",
		Header: map[string]string{
			"Content-type": "application/json; charset=UTF-8",
			"Origin":       "https://www.ingress.com",
			"Referer":      "https://www.ingress.com/intel",
		},
		Body: bytes.NewReader(Data),
	})
	if err != nil {
		return _json, err
	}
	if err := json.Unmarshal(response.Body, &_json); err != nil {
		return _json, err
	}
	return _json, nil
}

func (I *Ingress) send_msg(msg string) (_json Json, err error) {
	Data, err := json.Marshal(map[string]interface{}{
		"message": msg,
		"latE6":   I.Config.Get("latE6").(float64),
		"lngE6":   I.Config.Get("lngE6").(float64),
		"tab":     "faction",
		"v":       I.Config.Get("v").(string),
	})
	if err != nil {
		return _json, err
	}
	response, err := I.Request(&Options{
		Method: "POST",
		Url:    "https://www.ingress.com/r/sendPlext",
		Header: map[string]string{
			"Content-type": "application/json; charset=UTF-8",
			"Origin":       "https://www.ingress.com",
			"Referer":      "https://www.ingress.com/intel",
		},
		Body: bytes.NewReader(Data),
	})
	if err != nil {
		return _json, err
	}
	if err := json.Unmarshal(response.Body, &_json); err != nil {
		return _json, err
	}
	return _json, nil
}

func (I *Ingress) _reload() (err error) {
	var v string
	v, err = I.__get_user_v()
	if err != nil {
		return err
	}
	I.Config.Set("v", v)
	I.Header["X-CSRFToken"], err = I.__get_token()
	if err != nil {
		return err
	}
	return err
}

func (I *Ingress) auto_send_msg_new_agent() string {
	_j, err := I.get_msg()
	if err != nil {
		return "Failed to get the message"
	}
	_new_agent := Set.New()
	if _j["result"] == nil {
		return "Result is Empty"
	}
	for _, v := range _j["result"].([]interface{}) {
		if newAgent, ok := I.__check_new_agent(v.([]interface{})[2].(map[string]interface{})["plext"].(map[string]interface{})["text"].(string)); ok {
			_new_agent.Add(newAgent)
		}
	}
	if _new_agent.Len() == 0 {
		return "Not New Agent"
	}
	agents := ""
	values := []interface{}{}
	for _, v := range _new_agent.List() {
		values = append(values, "(\""+v+"\",\""+strconv.FormatInt(time.Now().Unix(), 10)+"\")")
		agents += "@" + v + " "
	}
	if res, err := I.send_msg(agents + I.__rand_msg()); err == nil {
		if res["result"].(string) == "success" {
			if _, err := I.Sqlite3.Exec("INSERT INTO `user` (`agent`, `createtime`) VALUES " + I.__join(values, ",")); err != nil {
				return "message send success,Info storage error"
			}
			return "message send success,Info storage success"
		}
	}
	return "Send Message Error"
}

func main() {
	_time := flag.Int("time", 16, "msg time")
	_sleep_time := flag.Int("sleep", 90, "msg time")
	flag.Parse()
	ingress, err := New(*_time)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(ingress.auto_send_msg_new_agent())
	limiter := time.Tick(time.Millisecond * 1000 * time.Duration(*_sleep_time))
	for true {
		<-limiter
		err = ingress._reload()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(ingress.auto_send_msg_new_agent())
	}
}

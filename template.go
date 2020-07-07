package beegoroutable

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"
)

// key is name, value is type, like publish_id => int64
var paramTypeMapping = map[string]string{}

func SetGlobalParamTypeMapping(m string) {
	ss := strings.Split(m, ";")
	for _, s := range ss {
		if s == "" {
			continue
		}
		m := strings.Split(s, ":")
		if len(m) != 2 {
			panic(`error param type mapping, must like "publish_id,app_id:int64,env:string"`)
		}

		ps := strings.Split(m[0], ",")
		for _, p := range ps {
			paramTypeMapping[strings.TrimPrefix(p, " ")] = strings.TrimPrefix(m[1], " ")
		}
	}
}

type Param struct {
	Name         string
	Type         string
	OriginalName string
}

var bodyParam = Param{
	Name: "body",
	Type: "interface{}",
}

type Api struct {
	Name       string
	Path       string
	PathParams []Param
	Params     []Param
	Method     string
	Body       string
}

func legalVarName(str string) string {

	if str == "type" {
		str = "typ"
	}

	if str == "body" {
		return "body_"
	}
	return snakeToCam(str)
}

func snakeToCam(str string) string {
	ss := strings.Split(strings.TrimPrefix(str, "_"), "_")

	var newSs []string
	newSs = append(newSs, ss[0])
	for _, s := range ss[1:] {
		if s == "" {
			continue
		}
		newSs = append(newSs, strings.ToUpper(s[:1])+s[1:])
	}
	return strings.Join(newSs, "")
}

func getParamType(n string) string {
	if t, ok := paramTypeMapping[n]; ok {
		return t
	}
	return "interface{}"
}

func (a *Api) parse() {

	var pathItems []string
	for _, v := range strings.Split(a.Path, "/") {
		if !strings.HasPrefix(v, ":") {
			pathItems = append(pathItems, v)
			continue
		}
		pathItems = append(pathItems, "%v")

		param := Param{
			Name:         legalVarName(v[1:]),
			OriginalName: v[1:],
			Type:         getParamType(v[1:]),
		}
		a.PathParams = append(a.PathParams, param)
	}
	a.Path = strings.Join(pathItems, "/")
	a.Path = "/" + strings.TrimPrefix(a.Path, "/")
	a.Params = append(a.Params, a.PathParams...)
	switch a.Method {
	case "PUT", "POST":
		a.Body = "body"
		a.Params = append(a.Params, bodyParam)
	default:
		a.Body = "nil"
	}
}

func ShowParams(apis []Api) {
	for i, _ := range apis {
		apis[i].parse()
		for _, p := range apis[i].PathParams {
			fmt.Println(apis[i].Method, apis[i].Path, p.OriginalName)
		}
	}
}

func GenerateCode(pkgName, clientName string, apis []Api) (string, error) {

	for i, _ := range apis {
		apis[i].parse()
	}

	data := map[string]interface{}{
		"Pkg":        pkgName,
		"ClientName": clientName,
		"Apis":       apis,
	}
	code, err := doGen(data)
	if err != nil {
		return "", nil
	}

	code, err = format.Source(code)
	if err != nil {
		return "", err
	}
	return string(code), nil
}

func doGen(api interface{}) ([]byte, error) {

	tpl := template.New("gotpl")
	tpl, err := tpl.Parse(tplText)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, api); err != nil {
		return nil, nil
	}
	return buf.Bytes(), nil
}

const tplText = `// Code generated - DO NOT EDIT.
package {{.Pkg}}

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type (
	Client struct {
		Endpoint string
		*http.Client
	}

	ApiError struct {
		Code    int64  ` + "`json:\"code\"`" + `
		Message string ` + "`json:\"message\"`" + `
	}

	Result struct {
		ApiError
		Result interface{} ` + "`json:\"result\"`" + `
	}

	Param struct {
		Header http.Header
		Query  url.Values
	}
)

type {{.ClientName}} interface {
    {{range .Apis}}
    {{.Name}}(ctx context.Context, {{range $_, $i := .Params}}{{$i.Name}} {{$i.Type}},{{end}}result interface{}, params ...Param) error
    {{end}}
}

func (e *ApiError) Error() string {
	bs, _ := json.Marshal(e)
	return string(bs)
}

func New(endPoint string, client *http.Client) {{.ClientName}} {

	if client == nil {
		client = http.DefaultClient
	}
	return &Client{
		Endpoint: endPoint,
		Client:   client,
	}
}

func NewParam() *Param {
	return &Param{}
}

type keyAdder interface {
	Add(key, value string)
}

// drop last one if len of pairs is odd
func (p *Param) add(store keyAdder, pairs ...interface{}) {
	pairs = pairs[0 : len(pairs)/2*2]
	for i := 0; i < len(pairs); i += 2 {
		store.Add(fmt.Sprintf("%v", pairs[i]), fmt.Sprintf("%v", pairs[i+1]))
	}
}

func (p *Param) WithHeaders(pairs ...interface{}) *Param {
	if p.Header == nil {
		p.Header = http.Header{}
	}
	p.add(p.Header, pairs...)
	return p
}

func (p *Param) WithQueries(pairs ...interface{}) *Param {
	if p.Query == nil {
		p.Query = url.Values{}
	}
	p.add(p.Query, pairs...)
	return p
}

func encode(body interface{}) io.Reader {
	if body == nil {
		return nil
	}
	buf := bytes.Buffer{}
	_ = json.NewEncoder(&buf).Encode(body)
	return &buf
}

func (c *Client) do(req *http.Request, result interface{}, params ...Param) error {

	q := req.URL.Query()
	changed := false
	for _, p := range params {
		if p.Header != nil {
			for k, vs := range p.Header {
				for _, v := range vs {
					req.Header.Add(k, v)
				}
			}
		}
		for k, _ := range p.Query {
			q.Set(k, p.Query.Get(k))
			changed = true
		}
	}
	if changed {
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	res := Result{Result: result}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	if res.Code != 0 {
		return &ApiError{
			Code:    res.Code,
			Message: res.Message,
		}
	}
	return nil
}

{{range $api := .Apis}}
func (c *Client) {{$api.Name}}(ctx context.Context, {{range $_,$i := $api.Params}}{{$i.Name}} {{$i.Type}},{{end}} result interface{}, params ...Param) error {
	urlStr := c.Endpoint + fmt.Sprintf("{{$api.Path}}"{{range $__,$p := $api.PathParams}}, {{$p.Name}} {{end}})
	
	req, err := http.NewRequest("{{$api.Method}}", urlStr, encode({{$api.Body}}))
	if err != nil {
		return err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	return c.do(req, result, params...)
}
{{end}}
`

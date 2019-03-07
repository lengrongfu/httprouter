package httprouter

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)
// BodyParse 用于自定义解析请求参数
var BodyParse RequestBodyParse = &defaultBodyParse{}

type RequestBodyParse interface {
	Parse(r *http.Request) ([]byte,error)
}



type defaultBodyParse struct {}

func getMapRequest(r *http.Request) map[string]interface{} {
	bytes, e := ioutil.ReadAll(r.Body)
	if e != nil && e != io.EOF {
		panic(e)
	}
	//手动释放io
	r.Body.Close()
	m := make(map[string]interface{})
	e = json.Unmarshal(bytes, &m)
	if e != nil && e != io.EOF {
		panic(e)
	}
	return m
}

func (d *defaultBodyParse) Parse (r *http.Request) (bytes []byte, e error) {
	defer r.Body.Close()
	return ioutil.ReadAll(r.Body)
}

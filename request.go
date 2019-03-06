package httprouter

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func handleRequest(r *http.Request) map[string]interface{} {
	defer r.Body.Close()
	bytes, e := ioutil.ReadAll(r.Body)
	if e != nil && e != io.EOF {
		panic(e)
	}
	m := make(map[string]interface{})
	e = json.Unmarshal(bytes, &m)
	if e != nil && e != io.EOF {
		panic(e)
	}
	return m
}

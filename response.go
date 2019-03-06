package httprouter

import (
	"encoding/json"
	"net/http"
	"reflect"
)

func responseHandler(callResult []reflect.Value, w http.ResponseWriter) {
	if len(callResult) > 1 {
		panic(ErrReturnResultNum)
	}
	if len(callResult) == 0 {
		return
	}
	value := callResult[0]
	switch value.Type().Kind() {
	case reflect.String:
		w.Write([]byte(value.Interface().(string)))
		return
	case reflect.Struct:
		bytes, _ := json.Marshal(value.Interface())
		w.Write(bytes)
		return
	case reflect.Slice:
	case reflect.Map:
	case reflect.Array:
	}
}

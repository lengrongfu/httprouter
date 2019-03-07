package httprouter

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

func responseHandler(callResult []reflect.Value, w http.ResponseWriter) {
	if len(callResult) > 1 {
		panic(fmt.Errorf(ErrReturnResultNum.Error(),len(callResult)))
	}
	if len(callResult) == 0 {
		return
	}
	value := callResult[0]
	switch value.Type().Kind() {
	case reflect.String:
		w.Write([]byte(value.Interface().(string)))
		return
	default:
		w.Write(ValuePara(value))
		return
	}
}

func ValuePara(value reflect.Value) []byte {
	builder := strings.Builder{}
	builder.WriteByte('{')
	for i := 0; i < value.Type().NumField(); i++  {
		field := value.Type().Field(i)

		if !value.Field(i).CanInterface() {
			//字段外部不可访问
			continue
		}
		builder.WriteByte('"')
		builder.WriteString(field.Tag.Get("json"))
		builder.WriteByte('"')
		builder.WriteByte(':')
		switch value.Field(i).Kind() {
		case reflect.String:
			if value.Field(i).CanInterface() {
				builder.WriteByte('"')
				builder.WriteString(fmt.Sprint(value.Field(i).Interface()))
				builder.WriteByte('"')
			}
			break
		case reflect.Struct:
			builder.Write(ValuePara(value.Field(i)))
			break
		case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
			builder.WriteString(fmt.Sprint(value.Field(i).Interface()))
			break
		case reflect.Map:
			panic(fmt.Errorf(errors.New("not support kind,now kind is:%s").Error(),reflect.Map.String()))
			break
		default:
			panic(fmt.Errorf(errors.New("not support kind,now kind is:%s").Error(),reflect.Map.String()))
		}
		builder.WriteByte(',')
	}
	builder.WriteByte('}')
	str := builder.String()
	str = str[:len(str)-2] + "}"
	return []byte(str)
}

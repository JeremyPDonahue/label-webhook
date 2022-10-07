package envconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

/*
ignored:"true"
env:"ENVIRONMENT_VARIABLE"
default:"default value"
*/

type StructInfo struct {
	Name         string
	Alt          string
	Key          string
	Field        reflect.Value
	Tags         reflect.StructTag
	Type         reflect.Type
	DefaultValue interface{}
}

func GetStructInfo(spec interface{}) ([]StructInfo, error) {
	s := reflect.ValueOf(spec)

	if s.Kind() != reflect.Pointer {
		return []StructInfo{}, fmt.Errorf("getStructInfo() was sent a %s instead of a pointer to a struct.\n", s.Kind())
	}

	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return []StructInfo{}, fmt.Errorf("getStructInfo() was sent a %s instead of a struct.\n", s.Kind())
	}
	typeOfSpec := s.Type()

	infos := make([]StructInfo, 0, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		ftype := typeOfSpec.Field(i)

		ignored, _ := strconv.ParseBool(ftype.Tag.Get("ignored"))
		if !f.CanSet() || ignored {
			continue
		}

		for f.Kind() == reflect.Pointer {
			if f.IsNil() {
				if f.Type().Elem().Kind() != reflect.Struct {
					break
				}
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		info := StructInfo{
			Name:  ftype.Name,
			Alt:   strings.ToUpper(ftype.Tag.Get("env")),
			Key:   ftype.Name,
			Field: f,
			Tags:  ftype.Tag,
			Type:  ftype.Type,
		}
		if info.Alt != "" {
			info.Key = info.Alt
		}
		info.Key = strings.ToUpper(info.Key)
		if ftype.Tag.Get("default") != "" {
			v, err := typeConversion(ftype.Type.String(), ftype.Tag.Get("default"))
			if err != nil {
				return []StructInfo{}, err
			}
			info.DefaultValue = v
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func typeConversion(t, v string) (interface{}, error) {
	switch t {
	case "string":
		return v, nil
	case "int":
		return strconv.ParseInt(v, 10, 0)
	case "int8":
		return strconv.ParseInt(v, 10, 8)
	case "int16":
		return strconv.ParseInt(v, 10, 16)
	case "int32":
		return strconv.ParseInt(v, 10, 32)
	case "int64":
		return strconv.ParseInt(v, 10, 64)
	case "uint":
		return strconv.ParseUint(v, 10, 0)
	case "uint16":
		return strconv.ParseUint(v, 10, 16)
	case "uint32":
		return strconv.ParseUint(v, 10, 32)
	case "uint64":
		return strconv.ParseUint(v, 10, 64)
	case "float32":
		return strconv.ParseFloat(v, 32)
	case "float64":
		return strconv.ParseFloat(v, 64)
	case "complex64":
		return strconv.ParseComplex(v, 64)
	case "complex128":
		return strconv.ParseComplex(v, 128)
	case "bool":
		return strconv.ParseBool(v)
	}
	return nil, fmt.Errorf("Unable to identify type.")
}

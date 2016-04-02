package bencode

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

type Encoder struct {
	bytes.Buffer
}

func (encoder *Encoder) writeString(str string) {
	encoder.WriteString(strconv.Itoa(len(str)))
	encoder.WriteByte(':')
	encoder.WriteString(str)
}

func (encoder *Encoder) writeInt(v int64) {
	encoder.WriteByte('i')
	encoder.WriteString(strconv.FormatInt(v, 10))
	encoder.WriteByte('e')
}

func (encoder *Encoder) writeUint(v uint64) {
	encoder.WriteByte('i')
	encoder.WriteString(strconv.FormatUint(v, 10))
	encoder.WriteByte('e')
}

func (encoder *Encoder) writeList(list []interface{}) error {
	encoder.WriteByte('l')
	for _, v := range list {
		switch value := v.(type) {
		case string:
			encoder.writeString(value)
		case []interface{}:
			if err := encoder.writeList(value); err != nil {
				return err
			}
		case map[string]interface{}:
			if err := encoder.writeDictionary(value); err != nil {
				return err
			}
		case int, int8, int16, int32, int64:
			encoder.writeInt(reflect.ValueOf(value).Int())
		case uint, uint8, uint16, uint32, uint64:
			encoder.writeUint(reflect.ValueOf(value).Uint())
		default:
			return fmt.Errorf("becode type error")
		}
	}
	encoder.WriteByte('e')
	return nil
}

func (encoder *Encoder) writeDictionary(dict map[string]interface{}) error {

	list := make(sort.StringSlice, len(dict))
	i := 0
	for key := range dict {
		list[i] = key
		i++
	}
	list.Sort()

	encoder.WriteByte('d')
	for _, key := range list {
		encoder.writeString(key)
		switch value := dict[key].(type) {
		case string:
			encoder.writeString(value)
		case []interface{}:
			if err := encoder.writeList(value); err != nil {
				return err
			}
		case map[string]interface{}:
			if err := encoder.writeDictionary(value); err != nil {
				return err
			}
		case int, int8, int16, int32, int64:
			encoder.writeInt(reflect.ValueOf(value).Int())
		case uint, uint8, uint16, uint32, uint64:
			encoder.writeUint(reflect.ValueOf(value).Uint())
		case []map[string]interface{}:
			encoder.WriteByte('l')
			for _, v := range value {
				if err := encoder.writeDictionary(v); err != nil {
					return err
				}
			}
			encoder.WriteByte('e')
		default:
			return fmt.Errorf("becode type error")
		}
	}
	encoder.WriteByte('e')
	return nil
}

func Marshal(dict map[string]interface{}) ([]byte, error) {
	encoder := Encoder{}
	if err := encoder.writeDictionary(dict); err != nil {
		return encoder.Bytes(), err
	}
	return encoder.Bytes(), nil
}

package bencode

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Decoder struct {
	bufio.Reader
}

func (decoder *Decoder) getInteger() (interface{}, error) {
	strInteger, err := decoder.ReadSlice('e')
	if err != nil {
		return nil, err
	}

	if integer, err := strconv.ParseInt(string(strInteger[:len(strInteger)-1]), 10, 64); err == nil {
		return integer, nil
	} else {
		return nil, err
	}
}

func (decoder *Decoder) getString() (string, error) {
	strLen, err := decoder.ReadSlice(':')
	if err != nil {
		return "", err
	}

	length, err := strconv.ParseUint(string(strLen[:len(strLen)-1]), 10, 64)
	if err != nil {
		return "", nil
	}

	buffer := make([]byte, length)
	_, err = io.ReadFull(decoder, buffer)
	return string(buffer), err
}

func (decoder *Decoder) getList() ([]interface{}, error) {
	var (
		list []interface{}
		item interface{}
	)
	for {
		ch, err := decoder.ReadByte()
		if err != nil {
			return nil, err
		}

		switch ch {
		case 'i':
			item, err = decoder.getInteger()
		case 'l':
			item, err = decoder.getList()
		case 'd':
			item, err = decoder.getDictionary()
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if err = decoder.UnreadByte(); err != nil {
				return nil, err
			}
			item, err = decoder.getString()
		default:
			return nil, fmt.Errorf("becode string should begin with 'i','l','d','0~9'")
		}
		list = append(list, item)
	}

	if ch, err := decoder.ReadByte(); err != nil {
		return nil, err
	} else if ch != 'e' {
		return nil, fmt.Errorf("becode list should end with 'e'")
	}

	return list, nil
}

func (decoder *Decoder) getDictionary() (map[string]interface{}, error) {
	dict := make(map[string]interface{})
	var item interface{}
	for {
		key, err := decoder.getString()
		if err != nil {
			return nil, err
		}

		ch, err := decoder.ReadByte()
		if err != nil {
			return nil, err
		}

		switch ch {
		case 'i':
			item, err = decoder.getInteger()
		case 'l':
			item, err = decoder.getList()
		case 'd':
			item, err = decoder.getDictionary()
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if err := decoder.UnreadByte(); err != nil {
				return nil, err
			}
			item, err = decoder.getString()
		default:
			return nil, fmt.Errorf("becode string should begin with 'i','l','d','0~9'")
		}

		dict[key] = item
	}

	if ch, err := decoder.ReadByte(); err != nil {
		return nil, err
	} else if ch != 'e' {
		return nil, fmt.Errorf("becode dictionary should end with 'e'")
	}

	return dict, nil
}

func Unmarshal(reader io.Reader) (map[string]interface{}, error) {
	decoder := Decoder{*bufio.NewReader(reader)}
	if ch, err := decoder.ReadByte(); err != nil {
		return nil, err
	} else if ch != 'd' {
		return nil, fmt.Errorf("bencode data must begin with dictionary")
	}
	return decoder.getDictionary()
}

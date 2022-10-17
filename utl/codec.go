package utl

import (
	"errors"
	"github.com/golang/protobuf/proto"
)

type CodeC struct{}

func (c *CodeC) Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, errors.New("nil interface")
	} else if v1, ok1 := v.([]byte); ok1 {
		return v1, nil
	} else if v2, ok2 := v.(proto.Message); ok2 {
		return proto.Marshal(v2)
	}
	return nil, errors.New("invalid interface")
}
func (c *CodeC) Unmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return errors.New("empty data")
	} else if v1, ok1 := v.(*[]byte); ok1 {
		*v1 = data
		//copy(v1, data)
		return nil
	} else if v2, ok2 := v.(proto.Message); ok2 {
		return proto.Unmarshal(data, v2)
	}
	return errors.New("invalid interface")
}
func (c *CodeC) Name() string {
	return "chenfeng123.cn.codeC"
}


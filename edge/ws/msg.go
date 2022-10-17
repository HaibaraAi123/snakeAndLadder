package ws

type Msg struct {
	Type       Type   `json:"type,omitempty"`
	Seq        uint32 `json:"seq,omitempty"`
	IsNeedResp bool   `json:"isNeedResp"`
	State      uint8  `json:"state,omitempty"`
	Method     []byte `json:"method,omitempty"`
	Data       []byte `json:"data,omitempty"`
}

type Type uint8
type State uint8

const (
	_ = iota
	TypeHeatBreak
	TypeMsgRequest
	TypeMsgResponse
)
const (
	StateOk = iota
	StateErrMethod
)


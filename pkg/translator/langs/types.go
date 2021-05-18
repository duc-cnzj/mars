package langs

import "encoding/json"

type Lang map[string]string

func (l *Lang) Bytes() []byte {
	marshal, _ := json.Marshal(&l)

	return marshal
}

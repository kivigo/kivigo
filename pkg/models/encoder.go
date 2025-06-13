package models

type Encoder interface {
	Encode(value any) ([]byte, error)
	Decode(data []byte, value any) error
}

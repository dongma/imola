package compresser

type Compresser interface {
	Code() byte
	Compress(data []byte) ([]byte, error)
	UnCompress(data []byte) ([]byte, error)
}

type DoNothing struct {
}

func (d DoNothing) Code() byte {
	return 0
}

func (d DoNothing) Compress(data []byte) ([]byte, error) {
	return data, nil
}

func (d DoNothing) UnCompress(data []byte) ([]byte, error) {
	return data, nil
}

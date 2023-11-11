package crypt

type Encrypted struct {
	iv      string
	content string
}

func (e *Encrypted) GetIv() string {
	return e.iv
}

func (e *Encrypted) GetContent() string {
	return e.content
}

func NewEncrypted(iv string, content string) *Encrypted {
	return &Encrypted{iv, content}
}

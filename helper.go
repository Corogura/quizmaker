package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
)

func generatePath() string {
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.URLEncoding, buf)
	input := make([]byte, 8)
	rand.Read(input)
	encoder.Write(input)
	return buf.String()
}

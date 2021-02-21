package braid

import (
	"io"
)

func readUntil(r io.Reader, delim byte) ([]byte, error) {
	var line []byte
	b := make([]byte, 1)
	for {
		_, err := r.Read(b)
		if err != nil {
			return line, err
		}
		line = append(line, b[0])
		if b[0] == delim {
			return line, nil
		}
	}
}

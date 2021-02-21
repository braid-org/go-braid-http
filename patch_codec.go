package braid

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Patch struct {
	Name          string
	ContentRange  string
	ContentLength uint64
	ExtraHeaders  map[string]string
	Body          []byte
}

func (p Patch) MarshalRequest() ([]byte, error) {
	if p.ContentLength == 0 {
		p.ContentLength = uint64(len(p.Body))
	}

	var buf bytes.Buffer

	if p.Name != "" {
		buf.WriteString(fmt.Sprintf("Patch-Name: \"%v\"\n", p.Name))
	}
	if p.ContentRange != "" {
		buf.WriteString(fmt.Sprintf("Content-Range: %v\n", p.ContentRange))
	}
	buf.WriteString(fmt.Sprintf("Content-Length: %v\n", p.ContentLength))

	for header, value := range p.ExtraHeaders {
		buf.WriteString(fmt.Sprintf("%v: %v\n", header, value))
	}

	buf.WriteString("\n")

	_, err := io.CopyN(&buf, bytes.NewReader(p.Body), int64(p.ContentLength))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type patchReadState int

const (
	patchReadStateHeaders patchReadState = iota
	patchReadStateBody
	patchReadStateDone
)

func (p *Patch) UnmarshalRequest(r io.Reader) error {
	state := patchReadStateHeaders
	for {
		line, err := readUntil(r, '\n')
		if err == io.EOF {
			state = patchReadStateDone
		} else if err != nil {
			return err
		}

		switch state {
		case patchReadStateHeaders:
			if len(bytes.TrimSpace(line)) == 0 {
				state = patchReadStateBody
				continue
			}
			err := p.handleHeader(line)
			if err != nil {
				return err
			}

		case patchReadStateBody, patchReadStateDone:
			if len(bytes.TrimSpace(line)) == 0 {
				state = patchReadStateDone
				break
			}
			err := p.handleBody(line)
			if err != nil {
				return err
			}
		}

		if state == patchReadStateDone {
			break
		}
	}
	if uint64(len(p.Body)) < p.ContentLength {
		return errors.Errorf("bad content length (expected %v, got %v)", p.ContentLength, len(p.Body))
	}
	p.Body = p.Body[:p.ContentLength]
	return nil
}

func (p *Patch) handleHeader(line []byte) error {
	if len(bytes.TrimSpace(line)) == 0 {
		return nil
	}

	parts := bytes.SplitN(line, []byte(":"), 2)
	if len(parts) < 2 {
		return errors.Errorf("bad patch header: %v", string(line))
	}
	header, value := bytes.TrimSpace(parts[0]), bytes.TrimSpace(parts[1])
	switch strings.ToLower(string(header)) {
	case "patch-name":
		p.Name = string(bytes.Trim(value, `"`))

	case "content-range":
		p.ContentRange = string(value)

	case "content-length":
		contentLength, err := strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return err
		}
		p.ContentLength = contentLength

	default:
		if p.ExtraHeaders == nil {
			p.ExtraHeaders = make(map[string]string)
		}
		p.ExtraHeaders[string(header)] = string(value)
	}
	return nil
}

func (p *Patch) handleBody(line []byte) error {
	if len(bytes.TrimSpace(line)) == 0 {
		return nil
	}
	p.Body = append(p.Body, line...)
	return nil
}

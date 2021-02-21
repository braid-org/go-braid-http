package braid

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type PutRequest struct {
	ContentType string
	Accept      string
	Version     string
	Parents     []string
	Patches     []Patch
}

// Creates an *http.Request representing the given Tx that follows the Braid-HTTP
// specification for sending transactions/patches to peers.
func MakePutRequest(ctx context.Context, dialAddr string, opts PutRequest) (*http.Request, error) {
	var patchBytes [][]byte
	for _, patch := range opts.Patches {
		bs, err := patch.MarshalRequest()
		if err != nil {
			return nil, err
		}
		patchBytes = append(patchBytes, bs)
	}
	body := bytes.NewBuffer(bytes.Join(patchBytes, []byte("\n\n")))

	req, err := http.NewRequestWithContext(ctx, "PUT", dialAddr, body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Header.Set("Version", opts.Version)
	req.Header.Set("Content-Type", opts.ContentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%v", body.Len()))
	if opts.Accept != "" {
		req.Header.Set("Accept", opts.Accept)
	}
	req.Header.Set("Parents", strings.Join(opts.Parents, ","))
	req.Header.Set("Patches", fmt.Sprintf("%v", len(opts.Patches)))
	return req, nil
}

func ReadPutRequest(r *http.Request) (*PutRequest, error) {
	contentType := r.Header.Get("Content-Type")
	accept := r.Header.Get("Accept")
	version := r.Header.Get("Version")

	parents := strings.Split(r.Header.Get("Parents"), ",")
	for i := range parents {
		parents[i] = strings.TrimSpace(parents[i])
	}

	numPatchesStr := r.Header.Get("Patches")
	if strings.TrimSpace(numPatchesStr) == "" {
		return nil, errors.New("missing Patches header")
	}
	numPatches, err := strconv.ParseUint(numPatchesStr, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "bad Patches header")
	}

	reader := bufio.NewReader(r.Body)

	var patches []Patch
	for i := uint64(0); i < numPatches; i++ {
		var patch Patch
		err := patch.UnmarshalRequest(reader)
		if err != nil {
			return nil, err
		}
		patches = append(patches, patch)
	}

	return &PutRequest{
		ContentType: contentType,
		Accept:      accept,
		Version:     version,
		Parents:     parents,
		Patches:     patches,
	}, nil
}

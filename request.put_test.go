package braid

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPutRequestRoundTrip(t *testing.T) {
	patch1 := []byte(`{"asdf": "jkl;"}`)
	patch2 := []byte(`{"braid": "http", "oof": ["rab", "zab"]}`)

	expected := &PutRequest{
		ContentType: "application/json",
		Accept:      "application/json",
		Version:     "12345",
		Parents:     []string{"foo", "bar"},
		Patches: []Patch{
			{
				Name:          "patch-type-1",
				ContentRange:  "json [-0:-0]",
				ContentLength: uint64(len(patch1)),
				ExtraHeaders: map[string]string{
					"Quux":  "xyzzy",
					"Quack": "duck",
				},
				Body: patch1,
			},
			{
				Name:          "patch-type-2",
				ContentRange:  "json .foo.bar",
				ContentLength: uint64(len(patch2)),
				ExtraHeaders: map[string]string{
					"Encoding": "flarf",
					"Cache":    "zork",
				},
				Body: patch2,
			},
		},
	}

	req, err := MakePutRequest(context.Background(), "http://braid.org", *expected)
	require.NoError(t, err)

	got, err := ReadPutRequest(req)
	require.NoError(t, err)

	require.Equal(t, expected, got)
}

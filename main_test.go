package main

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var proxyUrls = []string{}

func TestNewDataClient(t *testing.T) {
	file, err := os.Open("proxy.txt")
	require.NoError(t, err)
	defer require.NoError(t, file.Close())

	t.Log("Testing Get")

}

func TestGet(t *testing.T) {
	t.Log("Testing Get")

}

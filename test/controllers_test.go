//go:build e2e

package test

import (
	"context"
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	var cancelFunc context.CancelFunc
	ctx, cancelFunc = context.WithTimeout(context.Background(), flags.Timeout)
	defer cancelFunc()
	os.Exit(m.Run())
}

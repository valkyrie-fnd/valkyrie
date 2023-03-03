package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/valkyrie-fnd/valkyrie/internal/testutils"
)

func TestMain(t *testing.T) {
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()
	tests := []struct {
		Name           string
		Args           []string
		ExpectedExit   int
		OutputContains string
	}{
		{
			"Checking version should work just fine",
			[]string{"-version"},
			0,
			"Version: devel",
		},
		{
			"Starting Valkyrie without test config",
			[]string{},
			1,
			"Failed to read config",
		},
		{
			"Starting Valkyrie with test config",
			[]string{"-config", "./configs/testdata/valkyrie_config.test.main.yml"},
			0,
			"Operator server listening on 'localhost:",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(tt *testing.T) {
			tt.Setenv("VALK_PROFILES", "testlog")
			p, _ := testutils.GetFreePort()
			o, _ := testutils.GetFreePort()
			tt.Setenv("PROVIDER_PORT", fmt.Sprintf("%d", p))
			tt.Setenv("OPERATOR_PORT", fmt.Sprintf("%d", o))
			flag.CommandLine = flag.NewFlagSet(test.Name, flag.ExitOnError)
			os.Args = append([]string{test.Name}, test.Args...)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			var buf bytes.Buffer
			log.Logger = log.Output(&buf)

			exitCode := mainReal(ctx, &buf)

			assert.Equal(tt, test.ExpectedExit, exitCode, "Invalid exit code")
			output := buf.String()
			assert.Contains(t, output, test.OutputContains, "Output of log did not contain expected output")
		})
	}
}

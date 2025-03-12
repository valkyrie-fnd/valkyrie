#!/bin/sh

go list -u -m "$(go list -m -f '{{.Indirect}} {{.}}' all | grep '^false' | cut -d ' ' -f2)" | grep '\['

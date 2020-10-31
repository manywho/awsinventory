package awsdata_test

import (
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func assertTestErrorWasLogged(t *testing.T, entries []log.Entry) {
	assertErrorWasLogged(t, entries, testError)
}

func assertErrorWasLogged(t *testing.T, entries []log.Entry, err error) {
	var pass bool
	for _, entry := range entries {
		if strings.Contains(entry.Message, err.Error()) {
			pass = true
		}
	}

	require.True(t, pass, "expected error was not logged")
}

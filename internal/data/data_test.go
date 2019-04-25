package data_test

import (
	"bufio"
	"bytes"
	"testing"

	. "github.com/itmecho/awsinventory/internal/data"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLoadExitsEarlyWhenRegionsIsEmpty(t *testing.T) {
	var output bytes.Buffer
	buf := bufio.NewWriter(&output)

	logger := logrus.New()
	logger.SetOutput(buf)

	d := New(logger, TestClients{})

	d.Load([]string{}, []string{})

	buf.Flush()

	require.Contains(t, output.String(), ErrNoRegions.Error())
}

func TestLoadExitsEarlyWhenServicesIsEmpty(t *testing.T) {
	var output bytes.Buffer
	buf := bufio.NewWriter(&output)

	logger := logrus.New()
	logger.SetOutput(buf)

	d := New(logger, TestClients{})

	d.Load([]string{"test-region"}, []string{})

	buf.Flush()

	require.Contains(t, output.String(), ErrNoServices.Error())
}

package data_test

import (
	"testing"

	. "github.com/itmecho/awsinventory/internal/data"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestLoadExitsEarlyWhenRegionsIsEmptyAndRegionalServicesAreIncluded(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{})

	d.Load([]string{}, []string{ServiceEC2})

	require.Contains(t, hook.LastEntry().Message, ErrNoRegions.Error())
}

func TestLoadExitsEarlyWhenServicesIsEmpty(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{})

	d.Load([]string{ValidRegions[0]}, []string{})

	require.Contains(t, hook.LastEntry().Message, ErrNoServices.Error())
}

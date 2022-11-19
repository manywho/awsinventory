package awsdata_test

import (
	"errors"
	"testing"

	. "github.com/sudoinclabs/awsinventory/internal/awsdata"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestLoadExitsEarlyWhenRegionsIsEmptyAndRegionalServicesAreIncluded(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{})

	d.Load([]string{}, []string{ServiceEC2}, nil)

	assertErrorWasLogged(t, hook.Entries, ErrNoRegions)
}

func TestLoadCatchesInvalidRegion(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{})

	d.Load([]string{"test-region"}, []string{}, nil)

	assertErrorWasLogged(t, hook.Entries, errors.New("invalid region: test-region"))
}

func TestLoadCatchesInvalidService(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	d := New(logger, TestClients{})

	d.Load([]string{DefaultRegion}, []string{"invalid-service"}, nil)

	assertErrorWasLogged(t, hook.Entries, errors.New("invalid service: invalid-service"))
}

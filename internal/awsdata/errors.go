package awsdata

import (
	"errors"
	"fmt"
)

var (
	// ErrNoRegions is logged when no regions are given to the Load method
	ErrNoRegions = errors.New("no regions specified")

	// ErrNoServices is logged when no services are given to the Load method
	ErrNoServices = errors.New("no services specified")
)

func newErrInvalidRegion(region string) error {
	return fmt.Errorf("invalid region: %s", region)
}

func newErrInvalidService(service string) error {
	return fmt.Errorf("invalid service: %s", service)
}

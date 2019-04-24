# AWS Inventory

[![Build Status](https://travis-ci.org/itmecho/awsinventory.svg?branch=master)](https://travis-ci.org/itmecho/awsinventory)
[![Go Report Card](https://goreportcard.com/badge/github.com/itmecho/awsinventory)](https://goreportcard.com/report/github.com/itmecho/awsinventory)

AWS Inventory is a command line tool written in Go to fetch data from AWS and use it to generate a FedRAMP compliant inventory of your assets

## Building
The provided `Makefile` has a build task to handle building the binary. Just run
```
make build
```

## Testing
The `Makefile` has 2 targets for testing, `test` and `coverage`, the latter being the default target. The `test` target is meant to be used when developing and you don't want to be slowed down by the coverage generator.
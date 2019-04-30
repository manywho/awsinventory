# AWS Inventory

[![Build Status](https://travis-ci.org/manywho/awsinventory.svg?branch=master)](https://travis-ci.org/manywho/awsinventory)
[![Go Report Card](https://goreportcard.com/badge/github.com/manywho/awsinventory)](https://goreportcard.com/report/github.com/manywho/awsinventory)
[![codecov](https://codecov.io/gh/manywho/awsinventory/branch/master/graph/badge.svg)](https://codecov.io/gh/manywho/awsinventory)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
![GitHub release](https://img.shields.io/github/release/manywho/awsinventory.svg)

AWS Inventory is a command line tool written in Go to fetch data from AWS and use it to generate a FedRAMP compliant inventory of your assets

## Usage

To use awsinventory, simply call the binary and pass any configuration flags

```sh
# Build an inventory of services in the EU London AWS region
./awsinventory --regions eu-west-2
```

## Flags

```
Usage of ./awsinventory:
  -l, --log-level string     set the level of log output (default "warning")
  -o, --output-file string   path to the output file (default "inventory.csv")
      --print-regions        prints the available AWS regions
  -r, --regions strings      regions to gather data from
  -s, --services strings     services to gather data from (default [ebs,ec2,elb,iam,rds,s3])
```

## Development

### Building
The provided `Makefile` has a build target to handle building the binary

```sh
# Build the binary in the current directory
make build
```

### Testing
The `Makefile` has 2 targets for local testing, `test` and `test-full`

```sh
# Run tests
# This is meant for rapid development
make test

# Run tests with coverage and race detection
# This target should be run before committing
make test-full
```
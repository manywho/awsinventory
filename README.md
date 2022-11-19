# AWS Inventory

[![Build Status](https://travis-ci.org/manywho/awsinventory.svg?branch=master)](https://travis-ci.org/manywho/awsinventory)
[![Go Report Card](https://goreportcard.com/badge/github.com/sudoinclabs/awsinventory)](https://goreportcard.com/report/github.com/sudoinclabs/awsinventory)
[![codecov](https://codecov.io/gh/manywho/awsinventory/branch/master/graph/badge.svg)](https://codecov.io/gh/manywho/awsinventory)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
![GitHub release](https://img.shields.io/github/release/manywho/awsinventory.svg)

AWS Inventory is a command line tool written in Go to fetch data from AWS and use it to generate a FedRAMP compliant inventory of your assets.

## FedRAMP Compliance
AWS Inventory aims to output a CSV in accordance to the [FedRAMP inventory template](https://www.fedramp.gov/assets/resources/templates/SSP-A13-FedRAMP-Integrated-Inventory-Workbook-Template.xlsx) found [here](https://www.fedramp.gov/templates/).

## Usage

To use awsinventory, simply download the [latest release](https://github.com/sudoinclabs/awsinventory/releases/latest) for your system, make the binary executable, then call it, passing any configuration flags. It uses the AWS SDK for Go to create a session based on the [default credential provider chain](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials), including `~/.aws/credentials` and `~/.aws/config`.

```sh
# Example for Linux 64-bit
wget -O awsinventory  https://github.com/sudoinclabs/awsinventory/releases/download/$VERSION/awsinventory-$VERSION-linux-amd64
chmod 700 awsinventory

# Build an inventory of services in the Europe (London) AWS region
./awsinventory --regions eu-west-2
```

## Flags

```
Usage of ./awsinventory:
  -l, --log-level string     set the level of log output (default "warning")
  -o, --output-file string   path to the output file (default "inventory.csv")
      --print-regions        prints the available AWS regions
  -r, --regions strings      regions to gather data from
  -s, --services strings     services to gather data from (default [cloudfront,codecommit,dynamodb,ebs,ec2,ecr,ecs,elasticache,elb,elbv2,es,iam,kms,lambda,rds,s3,sqs])
  -v, --version              prints the version information
```

## Development

### Building
The provided `Makefile` has a build target to handle building the binary.

```sh
# Build the binary in the current directory
make build
```

### Testing
The `Makefile` has 2 targets for local testing: `test` and `test-full`.

```sh
# Run tests
# This is meant for rapid development
make test

# Run tests with coverage and race detection
# This target should be run before committing
make test-full
```

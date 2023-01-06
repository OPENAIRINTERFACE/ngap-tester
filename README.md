<p align="center">
    <a href="https://github.com/openairinterface/ngap-tester/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-BSD3clause-blue.svg" alt="License"></a>
    <a href="https://github.com/openairinterface/ngap-tester/graphs/contributors"><img src="https://img.shields.io/github/contributors/openairinterface/ngap-tester" alt="GitHub contributors"></a>
    <a href="https://github.com/openairinterface/ngap-tester/commits/main"><img src="https://img.shields.io/github/last-commit/openairinterface/ngap-tester" alt="GitHub last commit"></a>
    <a href="https://github.com/openairinterface/ngap-tester/commits/main"><img src="https://img.shields.io/github/commit-activity/y/openairinterface/ngap-tester" alt="GitHub commit activity the past year"></a>
</p>

# Continuous Integration Status

![Legacy Build Status](https://github.com/openairinterface/ngap-tester/actions/workflows/build_legacy_gnbsim.yml/badge.svg?branch=main)
![NGAP-Tester Build Status](https://github.com/openairinterface/ngap-tester/actions/workflows/build_ngap_tester.yml/badge.svg?branch=main)
![Go Linter Status](https://github.com/openairinterface/ngap-tester/actions/workflows/golangci-lint.yml/badge.svg?branch=main)

# Introduction

`ngap-tester` implements a NR-UE(s) and gNodeB simulator for 5G Radio Access Networks (RAN).

It is used to validate the interfaces to an 5G Core Network.

`ngap-tester` is currently used as NGAP testing engine for the following projects:
- [Magma](https://github.com/magma/magma): An open-source software platform that gives
network operators an open, flexible and extendable mobile core network solution.
- [OAI-5GC](https://gitlab.eurecom.fr/oai/cn5g/oai-cn5g-fed): An open-source 5G Standalone
Core Network

# Architecture

Architecture is still under modification. 

# License

# Building

Locally:

```
$ go mod vendor
$ go build -mod=vendor
$ ./ngap-tester --help
```

# Running

## Example running with OAI 5G Core network

The ngaptester [default config file](https://github.com/OPENAIRINTERFACE/ngap-tester/blob/donotlookatthis-1/testscenario/config-default.yaml) is customized for running ngaptester with a local [docker deployment](https://github.com/OPENAIRINTERFACE/ngap-tester/blob/donotlookatthis-1/test/docker-compose-basic-nrf.yaml) of OAI core network.

### Deploying OAI CN

First you have to [provision](https://gitlab.eurecom.fr/oai/cn5g/oai-cn5g-fed/-/blob/master/docs/RETRIEVE_OFFICIAL_IMAGES.md) OAI CN docker images of network functions.

Then 

```
cd ngap-tester/test && docker-compose -f ./docker-compose-basic-nrf.yaml up
```

### Running ngaptester

```
$ ./ngap-tester run --tf ./testscenario/example-test-file
```

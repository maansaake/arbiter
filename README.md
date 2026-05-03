# Arbiter

[![Main branch protection](https://github.com/maansaake/arbiter/actions/workflows/main.yaml/badge.svg)](https://github.com/maansaake/arbiter/actions/workflows/main.yaml)
[![Code scanning](https://github.com/maansaake/arbiter/actions/workflows/code-scanning.yaml/badge.svg)](https://github.com/maansaake/arbiter/actions/workflows/code-scanning.yaml)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/maansaake/arbiter)](https://goreportcard.com/report/github.com/maansaake/arbiter)
![tag](https://img.shields.io/github/v/tag/maansaake/arbiter?label=latest%20version)

Arbiter is a system testing framework aimed at improving software testability, arbiter provides a rich and flexible framework.

Arbiter does not aim to know anything about the system under test (SUT) and does not make any assumptions nor does it have any preferences around technologies or protocols.

All you have to do in order to get started is implement a module.

## Writing modules

Modules in arbiter are concrete Golang implementations of how to perform operations towards a SUT. For example, an arbiter testing module for a REST API would include a bunch of different HTTP requests.

A testing module can be written either as a simple module, meaning arbiter does not know anything about how it interacts with the underlying SUT, or as a verbose module which exposes a set list of possible operations. A module can also (optionally) implement the config interface to expose input configuration as required.

Modules are registered with the arbiter manager, the root level component of the framework, which then makes them available in the executable. It is possible to enable/disable modules and tweak any configuration that the testing modules exposes through command line arguments. Command line arguments are automatically added for each module that is registered with the arbiter manager.

### Simple testing modules

A testing module that does not expose any operations is considered a simple module. A simple module must itself implement traffic generation, as arbiter is not informed of any specific operations.

### Verbose testing modules

Verbose testing modules expose a set of operations. The exposed operations are made available as configuration options in the executable, where the user is able to determine frequencies (calls per minute per operation) and (optionally) any input arguments. Verbose modules provide the user the freedom of specifying settings per operation when running their tests, and arbiter takes care of traffic generation.

## Reproducability with traffic models

In order to easily reproduce tests between software releases, arbiter supports writing traffic models. Traffic models let the user write test files that can be given to the arbiter executable. Traffic models are YAML files that specify how arbiter should behave when a test is started. The YAML file specifies which testing modules should be enabled and disabled, and settings for different operations.

Generate a blank traffic model by using the `--generate-traffic-model` argument.

## Reporting

Arbiter can generate a YAML report on completion. The YAML file summarizes the test result by stating:

- (Optional) Test name
- (Optional) SUT version
- Start datetime
- Duration
- Traffic model

# This is a Work In Progress (WIP)
# Konveyor CLI tool

A CLI for accessing all of the tools in the Konveyor community.

## Usage

To search the `PATH` for valid plugins:
```
$ konveyor plugin list
```

To execute a plugin:
```
$ konveyor <plugin-name> <arg-1> <arg-2> ...
```

## Development

### Prerequisites

- Golang 1.18 or above

### Steps

To build from source:
```
$ make build
```

To run build and tests:
```
$ make ci
```
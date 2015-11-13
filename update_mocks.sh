#! /bin/bash -e

mkdir -p \
  monitor/mock \
  state/mock

mockgen github.com/nanopack/yoke/state State,Store > state/mock/mock.go
mockgen github.com/nanopack/yoke/monitor Performer > monitor/mock/mock.go

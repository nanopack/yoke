#! /bin/bash -e

mkdir -p \
  monitor/mock \
  state/mock

mockgen github.com/nanobox-io/yoke/state State,Store > state/mock/mock.go
mockgen github.com/nanobox-io/yoke/monitor Performer > monitor/mock/mock.go

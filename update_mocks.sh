#! /bin/bash -e

mkdir -p \
  monitor/mock \
  state/mock

mockgen github.com/nanobox-io/yoke/state Store > state/mock/mock.go
mockgen github.com/nanobox-io/yoke/monitor Monitor,Candidate,Performer > monitor/mock/mock.go

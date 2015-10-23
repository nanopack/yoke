#! /bin/bash -e

mkdir -p \
  monitor/mock

mockgen github.com/nanobox-io/yoke/monitor Monitor,Candidate,Performer > monitor/mock/mock.go
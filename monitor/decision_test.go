// Copyright (c) 2015 Pagoda Box Inc
//
// This Source Code Form is subject to the terms of the Mozilla Public License, v.
// 2.0. If a copy of the MPL was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.
//

package monitor_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/nanobox-io/yoke/monitor"
	"github.com/nanobox-io/yoke/monitor/mock"
	"testing"
)

func TestPrimary(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_monitor.NewMockCandidate(ctrl)
	other := mock_monitor.NewMockCandidate(ctrl)
	arbiter := mock_monitor.NewMockMonitor(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("initialized", nil)
	me.EXPECT().GetRole().Return("primary", nil)
	me.EXPECT().SetDBRole("active")
	perform.EXPECT().TransitionToActive(me)

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestSecondary(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_monitor.NewMockCandidate(ctrl)
	other := mock_monitor.NewMockCandidate(ctrl)
	arbiter := mock_monitor.NewMockMonitor(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("initialized", nil)
	me.EXPECT().GetRole().Return("secondary", nil)
	me.EXPECT().SetDBRole("backup")
	perform.EXPECT().TransitionToBackupOf(me, other)

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestSingle(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_monitor.NewMockCandidate(ctrl)
	other := mock_monitor.NewMockCandidate(ctrl)
	arbiter := mock_monitor.NewMockMonitor(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("single", nil)
	me.EXPECT().SetDBRole("backup")
	perform.EXPECT().TransitionToBackupOf(me, other)

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestActive(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_monitor.NewMockCandidate(ctrl)
	other := mock_monitor.NewMockCandidate(ctrl)
	arbiter := mock_monitor.NewMockMonitor(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("active", nil)
	me.EXPECT().SetDBRole("backup")
	perform.EXPECT().TransitionToBackupOf(me, other)

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestBackup(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_monitor.NewMockCandidate(ctrl)
	other := mock_monitor.NewMockCandidate(ctrl)
	arbiter := mock_monitor.NewMockMonitor(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("backup", nil)
	me.EXPECT().SetDBRole("active")
	perform.EXPECT().TransitionToActive(me)

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestOtherDead(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_monitor.NewMockCandidate(ctrl)
	other := mock_monitor.NewMockCandidate(ctrl)
	bounce := mock_monitor.NewMockCandidate(ctrl)
	arbiter := mock_monitor.NewMockMonitor(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("", errors.New("dead"))
	arbiter.EXPECT().Bounce(other).Return(bounce)
	bounce.EXPECT().GetDBRole().Return("dead", nil)

	me.EXPECT().GetDBRole().Return("active", nil)

	me.EXPECT().SetDBRole("single")
	perform.EXPECT().TransitionToSingle(me)

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestOtherDeadButSingle(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_monitor.NewMockCandidate(ctrl)
	other := mock_monitor.NewMockCandidate(ctrl)
	bounce := mock_monitor.NewMockCandidate(ctrl)
	arbiter := mock_monitor.NewMockMonitor(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("", errors.New("dead"))
	arbiter.EXPECT().Bounce(other).Return(bounce)
	bounce.EXPECT().GetDBRole().Return("", errors.New("dead"))

	me.EXPECT().GetDBRole().Return("single", nil)

	monitor.NewDecider(me, other, arbiter, perform)
}

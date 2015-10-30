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
	"github.com/nanobox-io/yoke/state/mock"
	"testing"
)

func TestPrimary(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("initialized", nil)
	me.EXPECT().GetRole().Return("primary", nil)
	me.EXPECT().SetDBRole("active")
	perform.EXPECT().TransitionToActive()

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestSecondary(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("initialized", nil)
	me.EXPECT().GetRole().Return("secondary", nil)
	me.EXPECT().SetDBRole("backup")
	perform.EXPECT().TransitionToBackup()

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestSingle(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("single", nil)
	me.EXPECT().SetDBRole("backup")
	perform.EXPECT().TransitionToBackup()

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestActive(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("active", nil)
	me.EXPECT().SetDBRole("backup")
	perform.EXPECT().TransitionToBackup()

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestBackup(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("backup", nil)
	me.EXPECT().SetDBRole("active")
	perform.EXPECT().TransitionToActive()

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestOtherDead(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	bounce := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("", errors.New("dead"))
	other.EXPECT().Location().Return("127.0.0.1:1234")
	arbiter.EXPECT().Bounce("127.0.0.1:1234").Return(bounce)
	bounce.EXPECT().GetDBRole().Return("dead", nil)

	me.EXPECT().GetDBRole().Return("active", nil)

	me.EXPECT().SetDBRole("single")
	perform.EXPECT().TransitionToSingle()

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestOtherDeadButSingle(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	bounce := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("", errors.New("dead"))
	other.EXPECT().Location().Return("127.0.0.1:1234")
	arbiter.EXPECT().Bounce("127.0.0.1:1234").Return(bounce)
	bounce.EXPECT().GetDBRole().Return("", errors.New("dead"))

	me.EXPECT().GetDBRole().Return("single", nil)

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestOtherDeadBackup(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	bounce := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready()
	arbiter.EXPECT().Ready()

	other.EXPECT().GetDBRole().Return("", errors.New("dead"))
	other.EXPECT().Location().Return("127.0.0.1:1234")
	arbiter.EXPECT().Bounce("127.0.0.1:1234").Return(bounce)
	bounce.EXPECT().GetDBRole().Return("dead", nil)

	me.EXPECT().GetDBRole().Return("backup", nil)
	me.EXPECT().HasSynced().Return(true, nil)

	me.EXPECT().SetDBRole("single")
	perform.EXPECT().TransitionToSingle()

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestOtherDeadBackupNotSync(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	bounce := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready().Times(2)
	arbiter.EXPECT().Ready().Times(2)

	other.EXPECT().GetDBRole().Return("", errors.New("dead")).Times(2)
	other.EXPECT().Location().Return("127.0.0.1:1234").Times(2)
	arbiter.EXPECT().Bounce("127.0.0.1:1234").Return(bounce).Times(2)
	bounce.EXPECT().GetDBRole().Return("dead", nil).Times(2)

	me.EXPECT().GetDBRole().Return("backup", nil).Times(2)
	me.EXPECT().HasSynced().Return(false, nil)

	perform.EXPECT().Stop()

	me.EXPECT().HasSynced().Return(true, nil)

	me.EXPECT().SetDBRole("single")
	perform.EXPECT().TransitionToSingle()

	monitor.NewDecider(me, other, arbiter, perform)
}

func TestOtherTemporaryDead(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)
	bounce := mock_state.NewMockState(ctrl)
	arbiter := mock_state.NewMockState(ctrl)
	perform := mock_monitor.NewMockPerformer(ctrl)

	other.EXPECT().Ready().Times(2)
	arbiter.EXPECT().Ready().Times(2)

	other.EXPECT().GetDBRole().Return("", errors.New("dead")).Times(2)
	other.EXPECT().Location().Return("127.0.0.1:1234").Times(2)
	arbiter.EXPECT().Bounce("127.0.0.1:1234").Return(bounce).Times(2)

	bounce.EXPECT().GetDBRole().Return("", errors.New("dead"))

	me.EXPECT().GetDBRole().Return("active", nil)
	perform.EXPECT().Stop()

	bounce.EXPECT().GetDBRole().Return("dead", nil)
	me.EXPECT().GetDBRole().Return("single", nil)

	me.EXPECT().SetDBRole("single")
	perform.EXPECT().TransitionToSingle()

	monitor.NewDecider(me, other, arbiter, perform)
}

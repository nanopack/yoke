// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/nanobox-io/yoke/monitor (interfaces: Monitor,Candidate,Performer)

package mock_monitor

import (
	gomock "github.com/golang/mock/gomock"
	monitor "github.com/nanobox-io/yoke/monitor"
)

// Mock of Monitor interface
type MockMonitor struct {
	ctrl     *gomock.Controller
	recorder *_MockMonitorRecorder
}

// Recorder for MockMonitor (not exported)
type _MockMonitorRecorder struct {
	mock *MockMonitor
}

func NewMockMonitor(ctrl *gomock.Controller) *MockMonitor {
	mock := &MockMonitor{ctrl: ctrl}
	mock.recorder = &_MockMonitorRecorder{mock}
	return mock
}

func (_m *MockMonitor) EXPECT() *_MockMonitorRecorder {
	return _m.recorder
}

func (_m *MockMonitor) Bounce(_param0 monitor.Candidate) monitor.Candidate {
	ret := _m.ctrl.Call(_m, "Bounce", _param0)
	ret0, _ := ret[0].(monitor.Candidate)
	return ret0
}

func (_mr *_MockMonitorRecorder) Bounce(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Bounce", arg0)
}

func (_m *MockMonitor) GetRole() (string, error) {
	ret := _m.ctrl.Call(_m, "GetRole")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockMonitorRecorder) GetRole() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetRole")
}

func (_m *MockMonitor) Ready() {
	_m.ctrl.Call(_m, "Ready")
}

func (_mr *_MockMonitorRecorder) Ready() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Ready")
}

// Mock of Candidate interface
type MockCandidate struct {
	ctrl     *gomock.Controller
	recorder *_MockCandidateRecorder
}

// Recorder for MockCandidate (not exported)
type _MockCandidateRecorder struct {
	mock *MockCandidate
}

func NewMockCandidate(ctrl *gomock.Controller) *MockCandidate {
	mock := &MockCandidate{ctrl: ctrl}
	mock.recorder = &_MockCandidateRecorder{mock}
	return mock
}

func (_m *MockCandidate) EXPECT() *_MockCandidateRecorder {
	return _m.recorder
}

func (_m *MockCandidate) Bounce(_param0 monitor.Candidate) monitor.Candidate {
	ret := _m.ctrl.Call(_m, "Bounce", _param0)
	ret0, _ := ret[0].(monitor.Candidate)
	return ret0
}

func (_mr *_MockCandidateRecorder) Bounce(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Bounce", arg0)
}

func (_m *MockCandidate) GetDBRole() (string, error) {
	ret := _m.ctrl.Call(_m, "GetDBRole")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockCandidateRecorder) GetDBRole() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetDBRole")
}

func (_m *MockCandidate) GetRole() (string, error) {
	ret := _m.ctrl.Call(_m, "GetRole")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockCandidateRecorder) GetRole() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetRole")
}

func (_m *MockCandidate) HasSynced() (bool, error) {
	ret := _m.ctrl.Call(_m, "HasSynced")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockCandidateRecorder) HasSynced() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "HasSynced")
}

func (_m *MockCandidate) Ready() {
	_m.ctrl.Call(_m, "Ready")
}

func (_mr *_MockCandidateRecorder) Ready() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Ready")
}

func (_m *MockCandidate) SetDBRole(_param0 string) error {
	ret := _m.ctrl.Call(_m, "SetDBRole", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockCandidateRecorder) SetDBRole(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetDBRole", arg0)
}

// Mock of Performer interface
type MockPerformer struct {
	ctrl     *gomock.Controller
	recorder *_MockPerformerRecorder
}

// Recorder for MockPerformer (not exported)
type _MockPerformerRecorder struct {
	mock *MockPerformer
}

func NewMockPerformer(ctrl *gomock.Controller) *MockPerformer {
	mock := &MockPerformer{ctrl: ctrl}
	mock.recorder = &_MockPerformerRecorder{mock}
	return mock
}

func (_m *MockPerformer) EXPECT() *_MockPerformerRecorder {
	return _m.recorder
}

func (_m *MockPerformer) Stop() {
	_m.ctrl.Call(_m, "Stop")
}

func (_mr *_MockPerformerRecorder) Stop() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Stop")
}

func (_m *MockPerformer) TransitionToActive(_param0 monitor.Candidate) {
	_m.ctrl.Call(_m, "TransitionToActive", _param0)
}

func (_mr *_MockPerformerRecorder) TransitionToActive(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "TransitionToActive", arg0)
}

func (_m *MockPerformer) TransitionToBackupOf(_param0 monitor.Candidate, _param1 monitor.Candidate) {
	_m.ctrl.Call(_m, "TransitionToBackupOf", _param0, _param1)
}

func (_mr *_MockPerformerRecorder) TransitionToBackupOf(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "TransitionToBackupOf", arg0, arg1)
}

func (_m *MockPerformer) TransitionToSingle(_param0 monitor.Candidate) {
	_m.ctrl.Call(_m, "TransitionToSingle", _param0)
}

func (_mr *_MockPerformerRecorder) TransitionToSingle(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "TransitionToSingle", arg0)
}
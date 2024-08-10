// Code generated by MockGen. DO NOT EDIT.
// Source: ./waypoint_reader.go
//
// Generated by this command:
//
//	mockgen -typed=true -source=./waypoint_reader.go -destination=./waypoint_reader_mock.go -package=sync
//

// Package sync is a generated GoMock package.
package sync

import (
	context "context"
	reflect "reflect"

	heimdall "github.com/erigontech/erigon/polygon/heimdall"
	gomock "go.uber.org/mock/gomock"
)

// MockwaypointReader is a mock of waypointReader interface.
type MockwaypointReader struct {
	ctrl     *gomock.Controller
	recorder *MockwaypointReaderMockRecorder
}

// MockwaypointReaderMockRecorder is the mock recorder for MockwaypointReader.
type MockwaypointReaderMockRecorder struct {
	mock *MockwaypointReader
}

// NewMockwaypointReader creates a new mock instance.
func NewMockwaypointReader(ctrl *gomock.Controller) *MockwaypointReader {
	mock := &MockwaypointReader{ctrl: ctrl}
	mock.recorder = &MockwaypointReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockwaypointReader) EXPECT() *MockwaypointReaderMockRecorder {
	return m.recorder
}

// CheckpointsFromBlock mocks base method.
func (m *MockwaypointReader) CheckpointsFromBlock(ctx context.Context, startBlock uint64) (heimdall.Waypoints, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckpointsFromBlock", ctx, startBlock)
	ret0, _ := ret[0].(heimdall.Waypoints)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckpointsFromBlock indicates an expected call of CheckpointsFromBlock.
func (mr *MockwaypointReaderMockRecorder) CheckpointsFromBlock(ctx, startBlock any) *MockwaypointReaderCheckpointsFromBlockCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckpointsFromBlock", reflect.TypeOf((*MockwaypointReader)(nil).CheckpointsFromBlock), ctx, startBlock)
	return &MockwaypointReaderCheckpointsFromBlockCall{Call: call}
}

// MockwaypointReaderCheckpointsFromBlockCall wrap *gomock.Call
type MockwaypointReaderCheckpointsFromBlockCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockwaypointReaderCheckpointsFromBlockCall) Return(arg0 heimdall.Waypoints, arg1 error) *MockwaypointReaderCheckpointsFromBlockCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockwaypointReaderCheckpointsFromBlockCall) Do(f func(context.Context, uint64) (heimdall.Waypoints, error)) *MockwaypointReaderCheckpointsFromBlockCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockwaypointReaderCheckpointsFromBlockCall) DoAndReturn(f func(context.Context, uint64) (heimdall.Waypoints, error)) *MockwaypointReaderCheckpointsFromBlockCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MilestonesFromBlock mocks base method.
func (m *MockwaypointReader) MilestonesFromBlock(ctx context.Context, startBlock uint64) (heimdall.Waypoints, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MilestonesFromBlock", ctx, startBlock)
	ret0, _ := ret[0].(heimdall.Waypoints)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MilestonesFromBlock indicates an expected call of MilestonesFromBlock.
func (mr *MockwaypointReaderMockRecorder) MilestonesFromBlock(ctx, startBlock any) *MockwaypointReaderMilestonesFromBlockCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MilestonesFromBlock", reflect.TypeOf((*MockwaypointReader)(nil).MilestonesFromBlock), ctx, startBlock)
	return &MockwaypointReaderMilestonesFromBlockCall{Call: call}
}

// MockwaypointReaderMilestonesFromBlockCall wrap *gomock.Call
type MockwaypointReaderMilestonesFromBlockCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockwaypointReaderMilestonesFromBlockCall) Return(arg0 heimdall.Waypoints, arg1 error) *MockwaypointReaderMilestonesFromBlockCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockwaypointReaderMilestonesFromBlockCall) Do(f func(context.Context, uint64) (heimdall.Waypoints, error)) *MockwaypointReaderMilestonesFromBlockCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockwaypointReaderMilestonesFromBlockCall) DoAndReturn(f func(context.Context, uint64) (heimdall.Waypoints, error)) *MockwaypointReaderMilestonesFromBlockCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

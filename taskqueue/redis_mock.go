// Code generated by MockGen. DO NOT EDIT.
// Source: redis.go

// Package taskqueue is a generated GoMock package.
package taskqueue

import (
	context "context"
	reflect "reflect"

	v8 "github.com/go-redis/redis/v8"
	gomock "github.com/golang/mock/gomock"
)

// MockRedis is a mock of Redis interface.
type MockRedis struct {
	ctrl     *gomock.Controller
	recorder *MockRedisMockRecorder
}

// MockRedisMockRecorder is the mock recorder for MockRedis.
type MockRedisMockRecorder struct {
	mock *MockRedis
}

// NewMockRedis creates a new mock instance.
func NewMockRedis(ctrl *gomock.Controller) *MockRedis {
	mock := &MockRedis{ctrl: ctrl}
	mock.recorder = &MockRedisMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRedis) EXPECT() *MockRedisMockRecorder {
	return m.recorder
}

// Del mocks base method.
func (m *MockRedis) Del(ctx context.Context, keys ...string) *v8.IntCmd {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range keys {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Del", varargs...)
	ret0, _ := ret[0].(*v8.IntCmd)
	return ret0
}

// Del indicates an expected call of Del.
func (mr *MockRedisMockRecorder) Del(ctx interface{}, keys ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, keys...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del", reflect.TypeOf((*MockRedis)(nil).Del), varargs...)
}

// EvalSha mocks base method.
func (m *MockRedis) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *v8.Cmd {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, sha1, keys}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "EvalSha", varargs...)
	ret0, _ := ret[0].(*v8.Cmd)
	return ret0
}

// EvalSha indicates an expected call of EvalSha.
func (mr *MockRedisMockRecorder) EvalSha(ctx, sha1, keys interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, sha1, keys}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EvalSha", reflect.TypeOf((*MockRedis)(nil).EvalSha), varargs...)
}

// ScriptLoad mocks base method.
func (m *MockRedis) ScriptLoad(ctx context.Context, script string) *v8.StringCmd {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScriptLoad", ctx, script)
	ret0, _ := ret[0].(*v8.StringCmd)
	return ret0
}

// ScriptLoad indicates an expected call of ScriptLoad.
func (mr *MockRedisMockRecorder) ScriptLoad(ctx, script interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScriptLoad", reflect.TypeOf((*MockRedis)(nil).ScriptLoad), ctx, script)
}

// ZAdd mocks base method.
func (m *MockRedis) ZAdd(ctx context.Context, key string, members ...*v8.Z) *v8.IntCmd {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, key}
	for _, a := range members {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ZAdd", varargs...)
	ret0, _ := ret[0].(*v8.IntCmd)
	return ret0
}

// ZAdd indicates an expected call of ZAdd.
func (mr *MockRedisMockRecorder) ZAdd(ctx, key interface{}, members ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, key}, members...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ZAdd", reflect.TypeOf((*MockRedis)(nil).ZAdd), varargs...)
}

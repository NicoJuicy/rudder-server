// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rudderlabs/rudder-server/services/streammanager/lambda (interfaces: LambdaClient)

// Package mock_lambda is a generated GoMock package.
package mock_lambda

import (
	reflect "reflect"

	lambda "github.com/aws/aws-sdk-go/service/lambda"
	gomock "github.com/golang/mock/gomock"
)

// MockLambdaClient is a mock of LambdaClient interface.
type MockLambdaClient struct {
	ctrl     *gomock.Controller
	recorder *MockLambdaClientMockRecorder
}

// MockLambdaClientMockRecorder is the mock recorder for MockLambdaClient.
type MockLambdaClientMockRecorder struct {
	mock *MockLambdaClient
}

// NewMockLambdaClient creates a new mock instance.
func NewMockLambdaClient(ctrl *gomock.Controller) *MockLambdaClient {
	mock := &MockLambdaClient{ctrl: ctrl}
	mock.recorder = &MockLambdaClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLambdaClient) EXPECT() *MockLambdaClientMockRecorder {
	return m.recorder
}

// Invoke mocks base method.
func (m *MockLambdaClient) Invoke(arg0 *lambda.InvokeInput) (*lambda.InvokeOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Invoke", arg0)
	ret0, _ := ret[0].(*lambda.InvokeOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Invoke indicates an expected call of Invoke.
func (mr *MockLambdaClientMockRecorder) Invoke(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Invoke", reflect.TypeOf((*MockLambdaClient)(nil).Invoke), arg0)
}
// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/kazemisoroush/code-refactor-tool/pkg/analyzer (interfaces: CodeAnalyzer)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
)

// MockCodeAnalyzer is a mock of CodeAnalyzer interface.
type MockCodeAnalyzer struct {
	ctrl     *gomock.Controller
	recorder *MockCodeAnalyzerMockRecorder
}

// MockCodeAnalyzerMockRecorder is the mock recorder for MockCodeAnalyzer.
type MockCodeAnalyzerMockRecorder struct {
	mock *MockCodeAnalyzer
}

// NewMockCodeAnalyzer creates a new mock instance.
func NewMockCodeAnalyzer(ctrl *gomock.Controller) *MockCodeAnalyzer {
	mock := &MockCodeAnalyzer{ctrl: ctrl}
	mock.recorder = &MockCodeAnalyzerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCodeAnalyzer) EXPECT() *MockCodeAnalyzerMockRecorder {
	return m.recorder
}

// AnalyzeCode mocks base method.
func (m *MockCodeAnalyzer) AnalyzeCode(arg0 string) (models.AnalysisResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AnalyzeCode", arg0)
	ret0, _ := ret[0].(models.AnalysisResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AnalyzeCode indicates an expected call of AnalyzeCode.
func (mr *MockCodeAnalyzerMockRecorder) AnalyzeCode(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AnalyzeCode", reflect.TypeOf((*MockCodeAnalyzer)(nil).AnalyzeCode), arg0)
}

// ExtractMetrics mocks base method.
func (m *MockCodeAnalyzer) ExtractMetrics(arg0 models.AnalysisResult) (models.CodeMetrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExtractMetrics", arg0)
	ret0, _ := ret[0].(models.CodeMetrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExtractMetrics indicates an expected call of ExtractMetrics.
func (mr *MockCodeAnalyzerMockRecorder) ExtractMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExtractMetrics", reflect.TypeOf((*MockCodeAnalyzer)(nil).ExtractMetrics), arg0)
}

// GenerateReport mocks base method.
func (m *MockCodeAnalyzer) GenerateReport(arg0 models.CodeMetrics) models.Report {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateReport", arg0)
	ret0, _ := ret[0].(models.Report)
	return ret0
}

// GenerateReport indicates an expected call of GenerateReport.
func (mr *MockCodeAnalyzerMockRecorder) GenerateReport(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateReport", reflect.TypeOf((*MockCodeAnalyzer)(nil).GenerateReport), arg0)
}

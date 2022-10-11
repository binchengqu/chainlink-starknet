// Code generated by mockery v2.12.0. DO NOT EDIT.

package mocks

import (
	context "context"
	big "math/big"

	mock "github.com/stretchr/testify/mock"

	ocr2 "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2"

	starknet "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"

	testing "testing"
)

// OCR2Reader is an autogenerated mock type for the OCR2Reader type
type OCR2Reader struct {
	mock.Mock
}

// BaseReader provides a mock function with given fields:
func (_m *OCR2Reader) BaseReader() starknet.Reader {
	ret := _m.Called()

	var r0 starknet.Reader
	if rf, ok := ret.Get(0).(func() starknet.Reader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(starknet.Reader)
		}
	}

	return r0
}

// BillingDetails provides a mock function with given fields: _a0, _a1
func (_m *OCR2Reader) BillingDetails(_a0 context.Context, _a1 string) (ocr2.BillingDetails, error) {
	ret := _m.Called(_a0, _a1)

	var r0 ocr2.BillingDetails
	if rf, ok := ret.Get(0).(func(context.Context, string) ocr2.BillingDetails); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(ocr2.BillingDetails)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConfigFromEventAt provides a mock function with given fields: _a0, _a1, _a2
func (_m *OCR2Reader) ConfigFromEventAt(_a0 context.Context, _a1 string, _a2 uint64) (ocr2.ContractConfig, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 ocr2.ContractConfig
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64) ocr2.ContractConfig); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(ocr2.ContractConfig)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, uint64) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LatestConfigDetails provides a mock function with given fields: _a0, _a1
func (_m *OCR2Reader) LatestConfigDetails(_a0 context.Context, _a1 string) (ocr2.ContractConfigDetails, error) {
	ret := _m.Called(_a0, _a1)

	var r0 ocr2.ContractConfigDetails
	if rf, ok := ret.Get(0).(func(context.Context, string) ocr2.ContractConfigDetails); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(ocr2.ContractConfigDetails)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LatestRoundData provides a mock function with given fields: _a0, _a1
func (_m *OCR2Reader) LatestRoundData(_a0 context.Context, _a1 string) (ocr2.RoundData, error) {
	ret := _m.Called(_a0, _a1)

	var r0 ocr2.RoundData
	if rf, ok := ret.Get(0).(func(context.Context, string) ocr2.RoundData); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(ocr2.RoundData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LatestTransmissionDetails provides a mock function with given fields: _a0, _a1
func (_m *OCR2Reader) LatestTransmissionDetails(_a0 context.Context, _a1 string) (ocr2.TransmissionDetails, error) {
	ret := _m.Called(_a0, _a1)

	var r0 ocr2.TransmissionDetails
	if rf, ok := ret.Get(0).(func(context.Context, string) ocr2.TransmissionDetails); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(ocr2.TransmissionDetails)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LinkAvailableForPayment provides a mock function with given fields: _a0, _a1
func (_m *OCR2Reader) LinkAvailableForPayment(_a0 context.Context, _a1 string) (*big.Int, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *big.Int
	if rf, ok := ret.Get(0).(func(context.Context, string) *big.Int); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewTransmissionsFromEventsAt provides a mock function with given fields: _a0, _a1, _a2
func (_m *OCR2Reader) NewTransmissionsFromEventsAt(_a0 context.Context, _a1 string, _a2 uint64) ([]ocr2.NewTransmissionEvent, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 []ocr2.NewTransmissionEvent
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64) []ocr2.NewTransmissionEvent); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]ocr2.NewTransmissionEvent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, uint64) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewOCR2Reader creates a new instance of OCR2Reader. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewOCR2Reader(t testing.TB) *OCR2Reader {
	mock := &OCR2Reader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

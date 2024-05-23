// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import (
	context "context"
	big "math/big"

	felt "github.com/NethermindEth/juno/core/felt"

	mock "github.com/stretchr/testify/mock"

	starknet "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

// ERC20Reader is an autogenerated mock type for the ERC20Reader type
type ERC20Reader struct {
	mock.Mock
}

// BalanceOf provides a mock function with given fields: _a0, _a1
func (_m *ERC20Reader) BalanceOf(_a0 context.Context, _a1 *felt.Felt) (*big.Int, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *big.Int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *felt.Felt) (*big.Int, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *felt.Felt) *big.Int); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *felt.Felt) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BaseReader provides a mock function with given fields:
func (_m *ERC20Reader) BaseReader() starknet.Reader {
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

// Decimals provides a mock function with given fields: _a0
func (_m *ERC20Reader) Decimals(_a0 context.Context) (*big.Int, error) {
	ret := _m.Called(_a0)

	var r0 *big.Int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*big.Int, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *big.Int); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewERC20Reader interface {
	mock.TestingT
	Cleanup(func())
}

// NewERC20Reader creates a new instance of ERC20Reader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewERC20Reader(t mockConstructorTestingTNewERC20Reader) *ERC20Reader {
	mock := &ERC20Reader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
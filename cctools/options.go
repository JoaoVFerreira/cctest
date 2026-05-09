package cctools

import (
	"github.com/JoaoVFerreira/cctest"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// WithSetup runs fn every time cctest creates a new Context.
func WithSetup(fn func() error) cctest.Option {
	return func(config *cctest.Config) {
		if fn == nil {
			return
		}
		config.AddContextSetup(func(*cctest.Context) error {
			return fn()
		})
	}
}

// WithChaincode configures the chaincode invoked by cctools helpers.
func WithChaincode(cc shim.Chaincode) cctest.Option {
	return func(config *cctest.Config) {
		config.SetChaincode(cc)
	}
}

package adapter

import (
	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

func (b *Binder) DeleteBinding(bindingId string, boshVMs bosh.BoshVMs, manifest bosh.BoshManifest, requestParameters serviceadapter.RequestParameters) error {

	return nil
}

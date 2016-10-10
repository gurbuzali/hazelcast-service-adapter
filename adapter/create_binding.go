package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

func (b *Binder) CreateBinding(bindingId string, boshVMs bosh.BoshVMs, manifest bosh.BoshManifest, requestParams serviceadapter.RequestParameters) (serviceadapter.Binding, error) {
	//params := requestParams.ArbitraryParams()

	hazelcastHosts := boshVMs["hazelcast"]
	if len(hazelcastHosts) == 0 {
		b.StderrLogger.Println("no VMs for instance group hazelcast")
		return serviceadapter.Binding{}, errors.New("")
	}

	var hazelCastAddresses []interface{}
	for _, hazelcastHost := range hazelcastHosts {
		hazelCastAddresses = append(hazelCastAddresses, fmt.Sprintf("%s", hazelcastHost))
	}

	/*
	    var invalidParams []string
	    for paramKey, _ := range params {
		    if paramKey != "topic" {
			    invalidParams = append(invalidParams, paramKey)
		    }
	    }

	    if len(invalidParams) > 0 {
		    sort.Strings(invalidParams)
		    errorMessage := fmt.Sprintf("unsupported parameter(s) for this service: %s", strings.Join(invalidParams, ", "))
		    b.StderrLogger.Println(errorMessage)
		    return serviceadapter.Binding{}, errors.New(errorMessage)
	    }
	*/

	group := manifest.InstanceGroups[0]
	//b.StderrLogger.Printf("Group: %+v \n", group)
	job := group.Jobs[0]
	//b.StderrLogger.Printf("Job: %+v \n", job)
	props := job.Properties["hazelcast"]
	//b.StderrLogger.Printf("Props: %+v \n", props)

	hazelcastProperties := props.(map[interface{}]interface{})

	return serviceadapter.Binding{
		Credentials: map[string]interface{}{
			"members": hazelCastAddresses,
			"group_name": hazelcastProperties["group_name"].(string),
			"group_pass": hazelcastProperties["group_pass"].(string),
		},
	}, nil
}

//go:generate counterfeiter -o fake_command_runner/fake_command_runner.go . CommandRunner
type CommandRunner interface {
	Run(name string, arg ...string) ([]byte, []byte, error)
}

type ExternalCommandRunner struct{}

func (c ExternalCommandRunner) Run(name string, arg ...string) ([]byte, []byte, error) {
	cmd := exec.Command(name, arg...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.Output()
	return stdout, stderr.Bytes(), err
}

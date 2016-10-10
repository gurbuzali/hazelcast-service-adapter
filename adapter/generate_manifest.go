package adapter

import (
	"errors"
	"fmt"
	"strings"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

const OnlyStemcellAlias = "only-stemcell"

func defaultDeploymentInstanceGroupsToJobs() map[string][]string {
	return map[string][]string{
		"hazelcast":     []string{"hazelcast"},
		"mancenter":         []string{"mancenter"},
	}
}

func getGroupNameAndPassword(xml string) (string, string) {
	groupArray := strings.SplitAfter(xml, "<group>")
	groupNameArray := strings.SplitAfter(groupArray[1], "<name>")
	valueArray := strings.Split(groupNameArray[1], "</name>")
	groupName := valueArray[0]

	passwordArray := strings.SplitAfter(valueArray[1], "<password>")
	passwordValueArray := strings.Split(passwordArray[1], "</password>")
	groupPass := passwordValueArray[0]
	return groupName, groupPass
}

func getMancenterProperties(xml string) (string, string, string) {
	var isEnabled string
	var updateInterval string
	var value string
	manArray := strings.Split(xml, "management-center")
	if len(manArray) == 3 {
		enabledArray := strings.SplitAfter(manArray[1], "enabled=\"")
		if len(enabledArray) != 1 {
			isEnabled = strings.Split(enabledArray[1], "\"")[0]
		}
		updateArray := strings.SplitAfter(manArray[1], "update-interval=\"")
		if len(updateArray) != 1 {
			updateInterval = strings.Split(updateArray[1], "\"")[0]
		}
		valueArray := strings.SplitAfter(manArray[1], ">")
		value = strings.Split(valueArray[1], "<")[0]
	}
	return isEnabled, updateInterval, value
}

func (a *ManifestGenerator) GenerateManifest(serviceDeployment serviceadapter.ServiceDeployment,
servicePlan serviceadapter.Plan,
requestParams serviceadapter.RequestParameters,
previousManifest *bosh.BoshManifest,
previousPlan *serviceadapter.Plan,
) (bosh.BoshManifest, error) {

	a.StderrLogger.Println(servicePlan)

	var releases []bosh.Release

	var hazelcastXml string;
	//var mancenterEnabled string
	//var mancenterUpdateInterval string
	//var mancenterUrl string
	groupName := "dev"
	groupPass := "dev-pass"

	if requestParams["parameters"] != nil {
		parameters := requestParams["parameters"].(map[string]interface{})
		hazelcastXmlParam := parameters["hazelcast_xml"]
		if hazelcastXmlParam != nil {
			hazelcastXml = hazelcastXmlParam.(string)
			groupName, groupPass = getGroupNameAndPassword(hazelcastXml)
			//mancenterEnabled, mancenterUpdateInterval, mancenterUrl = getMancenterProperties(hazelcastXml)
		}
	}

	for _, serviceRelease := range serviceDeployment.Releases {
		releases = append(releases, bosh.Release{
			Name:    serviceRelease.Name,
			Version: serviceRelease.Version,
		})
	}

	deploymentInstanceGroupsToJobs := defaultDeploymentInstanceGroupsToJobs()

	err := checkInstanceGroupsPresent([]string{"hazelcast"}, servicePlan.InstanceGroups)
	if err != nil {
		a.StderrLogger.Println(err.Error())
		return bosh.BoshManifest{}, errors.New("Contact your operator, service configuration issue occurred")
	}

	instanceGroups, err := InstanceGroupMapper(servicePlan.InstanceGroups, serviceDeployment.Releases, OnlyStemcellAlias, deploymentInstanceGroupsToJobs)
	if err != nil {
		a.StderrLogger.Println(err.Error())
		return bosh.BoshManifest{}, errors.New("")
	}

	hazelcastBrokerInstanceGroup := &instanceGroups[0]

	if len(hazelcastBrokerInstanceGroup.Networks) != 1 {
		a.StderrLogger.Println(fmt.Sprintf("expected 1 network for %s, got %d", hazelcastBrokerInstanceGroup.Name, len(hazelcastBrokerInstanceGroup.Networks)))
		return bosh.BoshManifest{}, errors.New("")
	}

	//arbitraryParameters := requestParams.ArbitraryParams()

	if hazelcastBrokerJob, ok := getJobFromInstanceGroup("hazelcast", hazelcastBrokerInstanceGroup); ok {
		hazelcastBrokerJob.Properties = map[string]interface{}{
			"network":                      hazelcastBrokerInstanceGroup.Networks[0].Name,
			"hazelcast": map[string]interface{}{
				"hazelcast_xml": hazelcastXml,
				"group_name": groupName,
				"group_pass": groupPass,
				//"mancenter_enabled": mancenterEnabled,
				//"mancenter_update_interval": mancenterUpdateInterval,
				//"mancenter_url": mancenterUrl,
			},
		}
	}

	manifestProperties := map[string]interface{}{}

	var updateBlock = bosh.Update{
		Canaries:        1,
		MaxInFlight:     1,
		CanaryWatchTime: "30000-240000",
		UpdateWatchTime: "30000-240000",
		Serial:          boolPointer(false),
	}

	if servicePlan.Update != nil {
		updateBlock = bosh.Update{
			Canaries:        servicePlan.Update.Canaries,
			MaxInFlight:     servicePlan.Update.MaxInFlight,
			CanaryWatchTime: servicePlan.Update.CanaryWatchTime,
			UpdateWatchTime: servicePlan.Update.UpdateWatchTime,
			Serial:          servicePlan.Update.Serial,
		}
	}

	return bosh.BoshManifest{
		Name:     serviceDeployment.DeploymentName,
		Releases: releases,
		Stemcells: []bosh.Stemcell{{
			Alias:   OnlyStemcellAlias,
			OS:      serviceDeployment.Stemcell.OS,
			Version: serviceDeployment.Stemcell.Version,
		}},
		InstanceGroups: instanceGroups,
		Properties:     manifestProperties,
		Update:         updateBlock,
	}, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getPreviousManifestProperty(name string, manifest *bosh.BoshManifest) (interface{}, bool) {
	if manifest != nil {
		if val, ok := manifest.Properties["auto_create_topics"]; ok {
			return val, true
		}
	}
	return nil, false
}

func getJobFromInstanceGroup(name string, instanceGroup *bosh.InstanceGroup) (*bosh.Job, bool) {
	for index, job := range instanceGroup.Jobs {
		if job.Name == name {
			return &instanceGroup.Jobs[index], true
		}
	}
	return &bosh.Job{}, false
}

func instanceCounts(plan serviceadapter.Plan) map[string]int {
	val := map[string]int{}
	for _, instanceGroup := range plan.InstanceGroups {
		val[instanceGroup.Name] = instanceGroup.Instances
	}
	return val
}

func boolPointer(b bool) *bool {
	return &b
}

func checkInstanceGroupsPresent(names []string, instanceGroups []serviceadapter.InstanceGroup) error {
	var missingNames []string

	for _, name := range names {
		if !containsInstanceGroup(name, instanceGroups) {
			missingNames = append(missingNames, name)
		}
	}

	if len(missingNames) > 0 {
		return fmt.Errorf("Invalid instance group configuration: expected to find: '%s' in list: '%s'",
			strings.Join(missingNames, ", "),
			strings.Join(getInstanceGroupNames(instanceGroups), ", "))
	}
	return nil
}

func getInstanceGroupNames(instanceGroups []serviceadapter.InstanceGroup) []string {
	var instanceGroupNames []string
	for _, instanceGroup := range instanceGroups {
		instanceGroupNames = append(instanceGroupNames, instanceGroup.Name)
	}
	return instanceGroupNames
}

func containsInstanceGroup(name string, instanceGroups []serviceadapter.InstanceGroup) bool {
	for _, instanceGroup := range instanceGroups {
		if instanceGroup.Name == name {
			return true
		}
	}

	return false
}

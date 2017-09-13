package convert

import "github.com/rancher/metadata/content"

func init() {
	content.ObjectFactories[content.ContainerType] = NewContainerObject
	content.ObjectFactories[content.ServiceType] = NewServiceObject
	content.ObjectFactories[content.StackType] = NewStackObject
	content.ObjectFactories[content.NetworkType] = NewNetworkObject
	content.ObjectFactories[content.HostType] = NewHostObject
	content.ObjectFactories[content.EnvironmentType] = NewEnvironmentObject
}

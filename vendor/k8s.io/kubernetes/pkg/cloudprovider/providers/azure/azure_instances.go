/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/cloudprovider"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/types"
)

// NodeAddresses returns the addresses of the specified instance.
func (az *Cloud) NodeAddresses(ctx context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	addressGetter := func(nodeName types.NodeName) ([]v1.NodeAddress, error) {
		ip, publicIP, err := az.GetIPForMachineWithRetry(nodeName)
		if err != nil {
			glog.V(2).Infof("NodeAddresses(%s) abort backoff: %v", nodeName, err)
			return nil, err
		}

		addresses := []v1.NodeAddress{
			{Type: v1.NodeInternalIP, Address: ip},
			{Type: v1.NodeHostName, Address: string(name)},
		}
		if len(publicIP) > 0 {
			addresses = append(addresses, v1.NodeAddress{
				Type:    v1.NodeExternalIP,
				Address: publicIP,
			})
		}
		return addresses, nil
	}

	if az.UseInstanceMetadata {
		metadata, err := az.metadata.GetMetadata()
		if err != nil {
			return nil, err
		}

		if metadata.Compute == nil || metadata.Network == nil {
			return nil, fmt.Errorf("failure of getting instance metadata")
		}

		isLocalInstance, err := az.isCurrentInstance(name, metadata.Compute.Name)
		if err != nil {
			return nil, err
		}

		// Not local instance, get addresses from Azure ARM API.
		if !isLocalInstance {
			return addressGetter(name)
		}

		if len(metadata.Network.Interface) == 0 {
			return nil, fmt.Errorf("no interface is found for the instance")
		}

		// Use ip address got from instance metadata.
		ipAddress := metadata.Network.Interface[0]
		addresses := []v1.NodeAddress{
			{Type: v1.NodeHostName, Address: string(name)},
		}
		if len(ipAddress.IPV4.IPAddress) > 0 && len(ipAddress.IPV4.IPAddress[0].PrivateIP) > 0 {
			address := ipAddress.IPV4.IPAddress[0]
			addresses = append(addresses, v1.NodeAddress{
				Type:    v1.NodeInternalIP,
				Address: address.PrivateIP,
			})
			if len(address.PublicIP) > 0 {
				addresses = append(addresses, v1.NodeAddress{
					Type:    v1.NodeExternalIP,
					Address: address.PublicIP,
				})
			}
		}
		if len(ipAddress.IPV6.IPAddress) > 0 && len(ipAddress.IPV6.IPAddress[0].PrivateIP) > 0 {
			address := ipAddress.IPV6.IPAddress[0]
			addresses = append(addresses, v1.NodeAddress{
				Type:    v1.NodeInternalIP,
				Address: address.PrivateIP,
			})
			if len(address.PublicIP) > 0 {
				addresses = append(addresses, v1.NodeAddress{
					Type:    v1.NodeExternalIP,
					Address: address.PublicIP,
				})
			}
		}

		if len(addresses) == 1 {
			// No IP addresses is got from instance metadata service, clean up cache and report errors.
			az.metadata.imsCache.Delete(metadataCacheKey)
			return nil, fmt.Errorf("get empty IP addresses from instance metadata service")
		}

		return addresses, nil
	}

	return addressGetter(name)
}

// NodeAddressesByProviderID returns the node addresses of an instances with the specified unique providerID
// This method will not be called from the node that is requesting this ID. i.e. metadata service
// and other local methods cannot be used here
func (az *Cloud) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	name, err := az.vmSet.GetNodeNameByProviderID(providerID)
	if err != nil {
		return nil, err
	}

	return az.NodeAddresses(ctx, name)
}

// ExternalID returns the cloud provider ID of the specified instance (deprecated).
func (az *Cloud) ExternalID(ctx context.Context, name types.NodeName) (string, error) {
	return az.InstanceID(ctx, name)
}

// InstanceExistsByProviderID returns true if the instance with the given provider id still exists and is running.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
func (az *Cloud) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	name, err := az.vmSet.GetNodeNameByProviderID(providerID)
	if err != nil {
		return false, err
	}

	_, err = az.InstanceID(ctx, name)
	if err != nil {
		if err == cloudprovider.InstanceNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (az *Cloud) isCurrentInstance(name types.NodeName, metadataVMName string) (bool, error) {
	var err error
	nodeName := mapNodeNameToVMName(name)
	if az.VMType == vmTypeVMSS {
		// VMSS vmName is not same with hostname, use hostname instead.
		metadataVMName, err = os.Hostname()
		if err != nil {
			return false, err
		}
	}

	metadataVMName = strings.ToLower(metadataVMName)
	return (metadataVMName == nodeName), err
}

// InstanceID returns the cloud provider ID of the specified instance.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (az *Cloud) InstanceID(ctx context.Context, name types.NodeName) (string, error) {
	nodeName := mapNodeNameToVMName(name)

	if az.UseInstanceMetadata {
		metadata, err := az.metadata.GetMetadata()
		if err != nil {
			return "", err
		}

		if metadata.Compute == nil {
			return "", fmt.Errorf("failure of getting instance metadata")
		}

		isLocalInstance, err := az.isCurrentInstance(name, metadata.Compute.Name)
		if err != nil {
			return "", err
		}

		// Not local instance, get instanceID from Azure ARM API.
		if !isLocalInstance {
			return az.vmSet.GetInstanceIDByNodeName(nodeName)
		}

		// Compose instanceID based on nodeName for standard instance.
		if az.VMType == vmTypeStandard {
			return az.getStandardMachineID(nodeName), nil
		}

		// Get scale set name and instanceID from vmName for vmss.
		ssName, instanceID, err := extractVmssVMName(metadata.Compute.Name)
		if err != nil {
			if err == ErrorNotVmssInstance {
				// Compose machineID for standard Node.
				return az.getStandardMachineID(nodeName), nil
			}
			return "", err
		}
		// Compose instanceID based on ssName and instanceID for vmss instance.
		return az.getVmssMachineID(ssName, instanceID), nil
	}

	return az.vmSet.GetInstanceIDByNodeName(nodeName)
}

// InstanceTypeByProviderID returns the cloudprovider instance type of the node with the specified unique providerID
// This method will not be called from the node that is requesting this ID. i.e. metadata service
// and other local methods cannot be used here
func (az *Cloud) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	name, err := az.vmSet.GetNodeNameByProviderID(providerID)
	if err != nil {
		return "", err
	}

	return az.InstanceType(ctx, name)
}

// InstanceType returns the type of the specified instance.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
// (Implementer Note): This is used by kubelet. Kubelet will label the node. Real log from kubelet:
//       Adding node label from cloud provider: beta.kubernetes.io/instance-type=[value]
func (az *Cloud) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	if az.UseInstanceMetadata {
		metadata, err := az.metadata.GetMetadata()
		if err != nil {
			return "", err
		}

		if metadata.Compute == nil {
			return "", fmt.Errorf("failure of getting instance metadata")
		}

		isLocalInstance, err := az.isCurrentInstance(name, metadata.Compute.Name)
		if err != nil {
			return "", err
		}
		if isLocalInstance {
			if metadata.Compute.VMSize != "" {
				return metadata.Compute.VMSize, nil
			}
		}
	}

	return az.vmSet.GetInstanceTypeByNodeName(string(name))
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (az *Cloud) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return fmt.Errorf("not supported")
}

// CurrentNodeName returns the name of the node we are currently running on.
// On Azure this is the hostname, so we just return the hostname.
func (az *Cloud) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}

// mapNodeNameToVMName maps a k8s NodeName to an Azure VM Name
// This is a simple string cast.
func mapNodeNameToVMName(nodeName types.NodeName) string {
	return string(nodeName)
}

/*
Copyright 2018 The Kubernetes Authors.

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
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	computepreview "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-12-01/compute"
	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

// AttachDisk attaches a vhd to vm
// the vhd must exist, can be identified by diskName, diskURI, and lun.
func (ss *scaleSet) AttachDisk(isManagedDisk bool, diskName, diskURI string, nodeName types.NodeName, lun int32, cachingMode compute.CachingTypes) error {
	ssName, instanceID, vm, err := ss.getVmssVM(string(nodeName))
	if err != nil {
		return err
	}

	disks := *vm.StorageProfile.DataDisks
	if isManagedDisk {
		disks = append(disks,
			computepreview.DataDisk{
				Name:         &diskName,
				Lun:          &lun,
				Caching:      computepreview.CachingTypes(cachingMode),
				CreateOption: "attach",
				ManagedDisk: &computepreview.ManagedDiskParameters{
					ID: &diskURI,
				},
			})
	} else {
		disks = append(disks,
			computepreview.DataDisk{
				Name: &diskName,
				Vhd: &computepreview.VirtualHardDisk{
					URI: &diskURI,
				},
				Lun:          &lun,
				Caching:      computepreview.CachingTypes(cachingMode),
				CreateOption: "attach",
			})
	}
	vm.StorageProfile.DataDisks = &disks

	ctx, cancel := getContextWithCancel()
	defer cancel()

	// Invalidate the cache right after updating
	defer ss.vmssVMCache.Delete(ss.makeVmssVMName(ssName, instanceID))

	glog.V(2).Infof("azureDisk - update(%s): vm(%s) - attach disk(%s)", ss.resourceGroup, nodeName, diskName)
	_, err = ss.VirtualMachineScaleSetVMsClient.Update(ctx, ss.resourceGroup, ssName, instanceID, vm)
	if err != nil {
		detail := err.Error()
		if strings.Contains(detail, errLeaseFailed) || strings.Contains(detail, errDiskBlobNotFound) {
			// if lease cannot be acquired or disk not found, immediately detach the disk and return the original error
			glog.Infof("azureDisk - err %s, try detach disk(%s)", detail, diskName)
			ss.DetachDiskByName(diskName, diskURI, nodeName)
		}
	} else {
		glog.V(2).Infof("azureDisk - attach disk(%s) succeeded", diskName)
	}
	return err
}

// DetachDiskByName detaches a vhd from host
// the vhd can be identified by diskName or diskURI
func (ss *scaleSet) DetachDiskByName(diskName, diskURI string, nodeName types.NodeName) error {
	ssName, instanceID, vm, err := ss.getVmssVM(string(nodeName))
	if err != nil {
		return err
	}

	disks := *vm.StorageProfile.DataDisks
	bFoundDisk := false
	for i, disk := range disks {
		if disk.Lun != nil && (disk.Name != nil && diskName != "" && *disk.Name == diskName) ||
			(disk.Vhd != nil && disk.Vhd.URI != nil && diskURI != "" && *disk.Vhd.URI == diskURI) ||
			(disk.ManagedDisk != nil && diskURI != "" && *disk.ManagedDisk.ID == diskURI) {
			// found the disk
			glog.V(2).Infof("azureDisk - detach disk: name %q uri %q", diskName, diskURI)
			disks = append(disks[:i], disks[i+1:]...)
			bFoundDisk = true
			break
		}
	}

	if !bFoundDisk {
		return fmt.Errorf("detach azure disk failure, disk %s not found, diskURI: %s", diskName, diskURI)
	}

	vm.StorageProfile.DataDisks = &disks
	ctx, cancel := getContextWithCancel()
	defer cancel()

	// Invalidate the cache right after updating
	defer ss.vmssVMCache.Delete(ss.makeVmssVMName(ssName, instanceID))

	glog.V(2).Infof("azureDisk - update(%s): vm(%s) - detach disk(%s)", ss.resourceGroup, nodeName, diskName)
	_, err = ss.VirtualMachineScaleSetVMsClient.Update(ctx, ss.resourceGroup, ssName, instanceID, vm)
	if err != nil {
		glog.Errorf("azureDisk - detach disk(%s) from %s failed, err: %v", diskName, nodeName, err)
	} else {
		glog.V(2).Infof("azureDisk - detach disk(%s) succeeded", diskName)
	}

	return err
}

// GetDiskLun finds the lun on the host that the vhd is attached to, given a vhd's diskName and diskURI
func (ss *scaleSet) GetDiskLun(diskName, diskURI string, nodeName types.NodeName) (int32, error) {
	_, _, vm, err := ss.getVmssVM(string(nodeName))
	if err != nil {
		return -1, err
	}

	disks := *vm.StorageProfile.DataDisks
	for _, disk := range disks {
		if disk.Lun != nil && (disk.Name != nil && diskName != "" && *disk.Name == diskName) ||
			(disk.Vhd != nil && disk.Vhd.URI != nil && diskURI != "" && *disk.Vhd.URI == diskURI) ||
			(disk.ManagedDisk != nil && *disk.ManagedDisk.ID == diskURI) {
			// found the disk
			glog.V(4).Infof("azureDisk - find disk: lun %d name %q uri %q", *disk.Lun, diskName, diskURI)
			return *disk.Lun, nil
		}
	}
	return -1, fmt.Errorf("Cannot find Lun for disk %s", diskName)
}

// GetNextDiskLun searches all vhd attachment on the host and find unused lun
// return -1 if all luns are used
func (ss *scaleSet) GetNextDiskLun(nodeName types.NodeName) (int32, error) {
	_, _, vm, err := ss.getVmssVM(string(nodeName))
	if err != nil {
		return -1, err
	}

	used := make([]bool, maxLUN)
	disks := *vm.StorageProfile.DataDisks
	for _, disk := range disks {
		if disk.Lun != nil {
			used[*disk.Lun] = true
		}
	}
	for k, v := range used {
		if !v {
			return int32(k), nil
		}
	}
	return -1, fmt.Errorf("All Luns are used")
}

// DisksAreAttached checks if a list of volumes are attached to the node with the specified NodeName
func (ss *scaleSet) DisksAreAttached(diskNames []string, nodeName types.NodeName) (map[string]bool, error) {
	attached := make(map[string]bool)
	for _, diskName := range diskNames {
		attached[diskName] = false
	}

	_, _, vm, err := ss.getVmssVM(string(nodeName))
	if err != nil {
		if err == cloudprovider.InstanceNotFound {
			// if host doesn't exist, no need to detach
			glog.Warningf("azureDisk - Cannot find node %q, DisksAreAttached will assume disks %v are not attached to it.",
				nodeName, diskNames)
			return attached, nil
		}

		return attached, err
	}

	disks := *vm.StorageProfile.DataDisks
	for _, disk := range disks {
		for _, diskName := range diskNames {
			if disk.Name != nil && diskName != "" && *disk.Name == diskName {
				attached[diskName] = true
			}
		}
	}

	return attached, nil
}

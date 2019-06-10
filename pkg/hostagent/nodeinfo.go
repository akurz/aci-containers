
// Copyright 2017 Cisco Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Handlers for nodeinfo.

package hostagent

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	nodeinfov1 "github.com/noironetworks/aci-containers/pkg/nodeinfo/apis/aci.nodeinfo"
	nodeinfoclientset "github.com/noironetworks/aci-containers/pkg/nodeinfo/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

nodeInfoInstance := &nodeinfov1.NodeInfo{
	ObjectMeta: metav1.ObjectMeta{
		Name: agent.config.NodeName,
	},
	Spec: nodeinfov1.NodeinfoSpec{
		Nodename: agent.config.NodeName,
		Macaddress : agent.config.UplinkMacAdress,
	},
}

func (agent *HostAgent) InformNodeInfo(nodeInfoClient *nodeinfoclientset.CientSet) {
	result, err := nodeInfoClient.AciV1().nodeinfos(metav1.NamespaceAll).Create(nodeInfoInstance)
	if err == nil {
		agent.log.Debug("NodeInfo CR is created")
	} else if apierrors.IsAlreadyExists(err) {
		agent.log.Debug("Already exists: %#\n", result)
	} else {
		panic(err)
	}
}

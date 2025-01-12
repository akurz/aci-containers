/***
Copyright 2018 Cisco Systems Inc. All rights reserved.

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
// GBP ep inventory definitions

package gbpserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	osexec "os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/noironetworks/aci-containers/pkg/apicapi"
	"github.com/pkg/errors"
)

type gbpInvMo struct {
	gbpCommonMo
}

const (
	epInvURI     = "/InvUniverse/InvRemoteEndpointInventory/"
	subjRemoteEP = "InvRemoteInventoryEp"
	getVtepsPath = "/usr/local/bin/get_vteps.sh"
)

var InvDB = make(map[string]map[string]*gbpInvMo)

func ReadInvFile(vtep, file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errorf("Reading %s - %v", file, err)
		return
	}

	var moList []gbpInvMo

	err = json.Unmarshal(data, &moList)
	if err != nil {
		log.Infof("Decoding %s - %v", file, err)
		return
	}

	moMap := make(map[string]*gbpInvMo)
	for _, mo := range moList {
		mm := new(gbpInvMo)
		*mm = mo
		moMap[mo.URI] = mm
	}
	InvDB[vtep] = moMap
}

func InitInvDB() {
	// fetch VTEPs and init them
	vtepList, err := osexec.Command(getVtepsPath).Output()
	if err != nil {
		log.Errorf("Getting vteps: %v", err)
		return
	}

	vteps := strings.TrimPrefix(strings.TrimSuffix(string(vtepList), "\n"), "\n")
	for _, vtep := range strings.Split(vteps, "\n") {
		log.Infof("Adding VTEP %s", vtep)
		_, ok := InvDB[vtep]
		if !ok {
			InvDB[vtep] = make(map[string]*gbpInvMo)
		}
	}

	DoAll()
}

func (g *gbpInvMo) save(vtep string) {
	_, ok := InvDB[vtep]
	if !ok {
		InvDB[vtep] = make(map[string]*gbpInvMo)
	}

	db := InvDB[vtep]
	db[g.URI] = g
}

func getInvMo(vtep, uri string) *gbpInvMo {
	db := InvDB[vtep]
	if db != nil {
		return db[uri]
	}

	return nil
}

func removeInvMo(vtep, uri string) {
	db := InvDB[vtep]
	if db != nil {
		delete(db, uri)
		log.Infof("Deleted %s, %s", uri, vtep)
	}
}

func GetInvMoMap(vtep string) map[string]*gbpCommonMo {
	res := make(map[string]*gbpCommonMo)
	for k, m := range InvDB {
		if k == vtep {
			continue // skip this vtep
		}

		for kk, mo := range m {
			res[kk] = &mo.gbpCommonMo
		}
	}

	return res
}

type Endpoint struct {
	Uuid    string `json:"uuid,omitempty"`
	MacAddr string `json:"macaddr,omitempty"`
	IPAddr  string `json:"ipaddr,omitempty"`
	EPG     string `json:"epg,omitempty"`
	VTEP    string `json:"vtep,omitempty"`
}

func (ep *Endpoint) Add() (string, error) {
	createChild := func(p *gbpCommonMo, childSub, name string) *gbpInvMo {
		var cURI string
		if name == "" {
			cURI = fmt.Sprintf("%s%s/", p.URI, childSub)
		} else {
			cURI = fmt.Sprintf("%s%s/%s/", p.URI, childSub, name)
		}
		child := &gbpInvMo{
			gbpCommonMo{
				Subject: childSub,
				URI:     cURI,
			},
		}
		child.SetParent(p.Subject, childSub, p.URI)
		child.save(ep.VTEP)
		p.AddChild(child.URI)
		return child
	}

	invMo := MoDB[epInvURI]
	if invMo == nil {
		return "", fmt.Errorf("epInventory not found")
	}
	// if it already exists, delete it from the tree
	epURI := fmt.Sprintf("%sInvRemoteInventoryEp/%s/", epInvURI, ep.Uuid)
	invMo.DelChild(epURI)

	epMo := createChild(&invMo.gbpCommonMo, subjRemoteEP, ep.Uuid)

	props := []Property{
		{Name: "mac", Data: ep.MacAddr},
		{Name: "nextHopTunnel", Data: ep.VTEP},
		{Name: "uuid", Data: ep.Uuid},
	}

	epMo.Properties = props

	ipMo := createChild(&epMo.gbpCommonMo, "InvRemoteIp", ep.IPAddr)
	ipMo.AddProperty("ip", ep.IPAddr)

	epgRefMo := createChild(&epMo.gbpCommonMo, "InvRemoteInventoryEpToGroupRSrc", "")
	epgURI := fmt.Sprintf("/PolicyUniverse/PolicySpace/%s/GbpEpGroup/%s/", kubeTenant, strings.Replace(ep.EPG, "|", "%7c", -1))
	ref := RefProperty{
		Subject: "GbpEpGroup",
		RefURI:  epgURI,
	}

	epgRefMo.AddProperty("target", ref)

	return epMo.URI, ep.pushTocAPIC(true)
}

func (ep *Endpoint) pushTocAPIC(add bool) error {
	if apicCon == nil {
		return nil
	}

	epToSg := apicapi.EmptyApicObject("hcloudRsEpToSecurityGroup", "")
	epToSg["hcloudRsEpToSecurityGroup"].Attributes["tDn"] = getSgDn(cApicName(ep.EPG))
	cEP := apicapi.EmptyApicObject("hcloudEndPoint", "")
	epName := string(ep.Uuid[len(ep.Uuid)-12:])
	epName = fmt.Sprintf("%s.%s", epName, ep.VTEP)
	cEP["hcloudEndPoint"].Attributes["name"] = epName
	cEP["hcloudEndPoint"].Attributes["primaryIpV4Addr"] = ep.IPAddr
	if !add {
		cEP["hcloudEndPoint"].Attributes["status"] = "deleted"
	}
	cEP["hcloudEndPoint"].Children = append(cEP["hcloudEndPoint"].Children, epToSg)

	cSN := apicapi.EmptyApicObject("hcloudSubnet", "")
	cSN["hcloudSubnet"].Attributes["addr"] = defCAPICSubnet
	cSN["hcloudSubnet"].Children = append(cSN["hcloudSubnet"].Children, cEP)

	cCidr := apicapi.EmptyApicObject("hcloudCidr", "")
	cCidr["hcloudCidr"].Attributes["addr"] = defCAPICCidr
	cCidr["hcloudCidr"].Children = append(cCidr["hcloudCidr"].Children, cSN)

	cCtx := apicapi.EmptyApicObject("hcloudCtx", "")
	cCtx["hcloudCtx"].Attributes["name"] = defVrfName
	cCtx["hcloudCtx"].Attributes["primaryCidr"] = defCAPICCidr
	cCtx["hcloudCtx"].Children = append(cCtx["hcloudCtx"].Children, cCidr)

	cRegion := apicapi.EmptyApicObject("hcloudRegion", "")
	cRegion["hcloudRegion"].Attributes["regionName"] = defRegion
	cRegion["hcloudRegion"].Children = append(cRegion["hcloudRegion"].Children, cCtx)

	cAcc := apicapi.EmptyApicObject("hcloudAccount", "")
	cAcc["hcloudAccount"].Attributes["name"] = kubeTenant
	cAcc["hcloudAccount"].Children = append(cAcc["hcloudAccount"].Children, cRegion)

	err := apicCon.PostTestAPI(cAcc)
	log.Errorf("pushTocAPIC: %v", err)
	return err
}

func (ep *Endpoint) FromMo(mo *gbpInvMo) error {
	if mo.Subject != subjRemoteEP {
		return fmt.Errorf("Mo class %s is not remote EP", mo.Subject)
	}

	ep.MacAddr = mo.GetStringProperty("mac")
	ep.VTEP = mo.GetStringProperty("nextHopTunnel")
	ep.Uuid = mo.GetStringProperty("uuid")

	m := InvDB[ep.VTEP]

	for _, c := range mo.Children {
		cMo, ok := m[c]
		if !ok {
			return fmt.Errorf("Child %s not found", c)
		}

		if cMo.Subject == "InvRemoteIp" {
			ep.IPAddr = cMo.GetStringProperty("ip")
		}

		if cMo.Subject == "InvRemoteInventoryEpToGroupRSrc" {
			if len(cMo.Properties) != 1 {
				return fmt.Errorf("Bad refmo %s", c)
			}
			rp, ok := cMo.Properties[0].Data.(RefProperty)
			if !ok {
				return fmt.Errorf("Bad prop refmo %s", c)
			}

			epgURI := strings.Split(rp.RefURI, "/")
			if len(epgURI) < 6 {
				return fmt.Errorf("Malformed refuri %s", rp.RefURI)
			}
			ep.EPG = epgURI[5]
		}
	}

	return nil
}

func (ep *Endpoint) Delete() error {
	epURI := fmt.Sprintf("%sInvRemoteInventoryEp/%s/", epInvURI, ep.Uuid)
	epMo := getInvMo(ep.VTEP, epURI)
	if epMo == nil {
		return fmt.Errorf("%s Not found", epURI)
	}

	for _, u := range epMo.Children {
		removeInvMo(ep.VTEP, u)
	}
	removeInvMo(ep.VTEP, epURI)
	invMo := MoDB[epInvURI]
	if invMo == nil {
		return fmt.Errorf("epInventory not found")
	}
	invMo.DelChild(epURI)

	return ep.pushTocAPIC(false)
}

func getSgDn(epgName string) string {
	cepg := apicapi.NewCloudEpg(kubeTenant, defCloudApp, epgName)
	n := fmt.Sprintf("acct-[%s]/region-[%s]/context-[%s]/sgroup-[%s]",
		kubeTenant, defRegion, defVrfName, cepg.GetDn())
	return n
}

// postEndpoint rest handler to create an endpoint
func postEndpoint(w http.ResponseWriter, r *http.Request, vars map[string]string) (interface{}, error) {
	gMutex.Lock()
	defer gMutex.Unlock()

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadAll")
	}

	ep := &Endpoint{}
	err = json.Unmarshal(content, ep)
	if err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	uri, err := ep.Add()
	DoAll()
	return &PostResp{URI: uri}, err
}

func getAllEPs() map[string]*gbpInvMo {
	allEPs := make(map[string]*gbpInvMo)

	for _, m := range InvDB {
		for _, mo := range m {
			if mo.Subject == subjRemoteEP {
				allEPs[mo.URI] = mo
			}
		}
	}

	return allEPs
}

func listEndpoints(w http.ResponseWriter, r *http.Request, vars map[string]string) (interface{}, error) {
	gMutex.Lock()
	defer gMutex.Unlock()

	var resp ListResp

	allEPs := getAllEPs()
	for u := range allEPs {
		resp.URIs = append(resp.URIs, u)
	}

	return &resp, nil
}

func fetchEP(r *http.Request) (*Endpoint, error) {
	params := r.URL.Query()
	key, ok := params["key"]
	if !ok {
		return nil, fmt.Errorf("key is missing")
	}

	allEPs := getAllEPs()
	epMo, ok := allEPs[key[0]]
	if !ok {
		return nil, fmt.Errorf("Not found - %s", key)
	}

	ep := &Endpoint{}
	err := ep.FromMo(epMo)
	return ep, err
}
func getEndpoint(w http.ResponseWriter, r *http.Request, vars map[string]string) (interface{}, error) {
	gMutex.Lock()
	defer gMutex.Unlock()

	return fetchEP(r)
}
func deleteEndpoint(w http.ResponseWriter, r *http.Request, vars map[string]string) (interface{}, error) {
	gMutex.Lock()
	defer gMutex.Unlock()

	ep, err := fetchEP(r)
	if err != nil {
		return nil, err
	}

	log.Infof("deleteEndpoint - VTEP: %s", ep.VTEP)
	err = ep.Delete()
	DoAll()
	return nil, err
}

/*
Copyright 2019 The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package internalversion

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	rbac "k8s.io/kubernetes/pkg/apis/rbac"
	scheme "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/scheme"
)

// ClusterRolesGetter has a method to return a ClusterRoleInterface.
// A group's client should implement this interface.
type ClusterRolesGetter interface {
	ClusterRoles() ClusterRoleInterface
}

// ClusterRoleInterface has methods to work with ClusterRole resources.
type ClusterRoleInterface interface {
	Create(*rbac.ClusterRole) (*rbac.ClusterRole, error)
	Update(*rbac.ClusterRole) (*rbac.ClusterRole, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*rbac.ClusterRole, error)
	List(opts v1.ListOptions) (*rbac.ClusterRoleList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *rbac.ClusterRole, err error)
	ClusterRoleExpansion
}

// clusterRoles implements ClusterRoleInterface
type clusterRoles struct {
	client rest.Interface
}

// newClusterRoles returns a ClusterRoles
func newClusterRoles(c *RbacClient) *clusterRoles {
	return &clusterRoles{
		client: c.RESTClient(),
	}
}

// Get takes name of the clusterRole, and returns the corresponding clusterRole object, and an error if there is any.
func (c *clusterRoles) Get(name string, options v1.GetOptions) (result *rbac.ClusterRole, err error) {
	result = &rbac.ClusterRole{}
	err = c.client.Get().
		Resource("clusterroles").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ClusterRoles that match those selectors.
func (c *clusterRoles) List(opts v1.ListOptions) (result *rbac.ClusterRoleList, err error) {
	result = &rbac.ClusterRoleList{}
	err = c.client.Get().
		Resource("clusterroles").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested clusterRoles.
func (c *clusterRoles) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("clusterroles").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a clusterRole and creates it.  Returns the server's representation of the clusterRole, and an error, if there is any.
func (c *clusterRoles) Create(clusterRole *rbac.ClusterRole) (result *rbac.ClusterRole, err error) {
	result = &rbac.ClusterRole{}
	err = c.client.Post().
		Resource("clusterroles").
		Body(clusterRole).
		Do().
		Into(result)
	return
}

// Update takes the representation of a clusterRole and updates it. Returns the server's representation of the clusterRole, and an error, if there is any.
func (c *clusterRoles) Update(clusterRole *rbac.ClusterRole) (result *rbac.ClusterRole, err error) {
	result = &rbac.ClusterRole{}
	err = c.client.Put().
		Resource("clusterroles").
		Name(clusterRole.Name).
		Body(clusterRole).
		Do().
		Into(result)
	return
}

// Delete takes name of the clusterRole and deletes it. Returns an error if one occurs.
func (c *clusterRoles) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("clusterroles").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *clusterRoles) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("clusterroles").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched clusterRole.
func (c *clusterRoles) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *rbac.ClusterRole, err error) {
	result = &rbac.ClusterRole{}
	err = c.client.Patch(pt).
		Resource("clusterroles").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}

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

// Code generated by informer-gen. DO NOT EDIT.

package internalversion

import (
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	storage "k8s.io/kubernetes/pkg/apis/storage"
	internalclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	internalinterfaces "k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion/internalinterfaces"
	internalversion "k8s.io/kubernetes/pkg/client/listers/storage/internalversion"
)

// VolumeAttachmentInformer provides access to a shared informer and lister for
// VolumeAttachments.
type VolumeAttachmentInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() internalversion.VolumeAttachmentLister
}

type volumeAttachmentInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewVolumeAttachmentInformer constructs a new informer for VolumeAttachment type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewVolumeAttachmentInformer(client internalclientset.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredVolumeAttachmentInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredVolumeAttachmentInformer constructs a new informer for VolumeAttachment type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredVolumeAttachmentInformer(client internalclientset.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Storage().VolumeAttachments().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Storage().VolumeAttachments().Watch(options)
			},
		},
		&storage.VolumeAttachment{},
		resyncPeriod,
		indexers,
	)
}

func (f *volumeAttachmentInformer) defaultInformer(client internalclientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredVolumeAttachmentInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *volumeAttachmentInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&storage.VolumeAttachment{}, f.defaultInformer)
}

func (f *volumeAttachmentInformer) Lister() internalversion.VolumeAttachmentLister {
	return internalversion.NewVolumeAttachmentLister(f.Informer().GetIndexer())
}

// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"io"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/cli-utils/pkg/inventory"
	"sigs.k8s.io/cli-utils/pkg/manifestreader"
)

// InventoryProvider implements the Provider interface.
var _ Provider = &InventoryProvider{}

// Provider is an interface which wraps the kubectl factory and
// the inventory client.
type Provider interface {
	Factory() util.Factory
	InventoryClient() (inventory.InventoryClient, error)
	InventoryInfo([]*unstructured.Unstructured) (inventory.InventoryInfo, []*unstructured.Unstructured, error)
	ToRESTMapper() (meta.RESTMapper, error)
	ManifestReader(reader io.Reader, args []string) (manifestreader.ManifestReader, error)
}

// InventoryProvider implements the Provider interface.
type InventoryProvider struct {
	factory util.Factory
}

// NewProvider returns a Provider that implements a ConfigMap inventory object.
func NewProvider(f util.Factory) *InventoryProvider {
	return &InventoryProvider{
		factory: f,
	}
}

// Factory returns the kubectl factory.
func (f *InventoryProvider) Factory() util.Factory {
	return f.factory
}

// InventoryClient returns an InventoryClient created with the stored
// factory and InventoryFactoryFunc values, or an error if one occurred.
func (f *InventoryProvider) InventoryClient() (inventory.InventoryClient, error) {
	return inventory.NewInventoryClient(f.factory,
		inventory.WrapInventoryObj,
		inventory.InvInfoToConfigMap,
	)
}

func (f *InventoryProvider) InventoryInfo(objs []*unstructured.Unstructured) (inventory.InventoryInfo, []*unstructured.Unstructured, error) {
	return inventory.SplitUnstructureds(objs)
}

// ToRESTMapper returns a RESTMapper created by the stored kubectl factory.
func (f *InventoryProvider) ToRESTMapper() (meta.RESTMapper, error) {
	return f.factory.ToRESTMapper()
}

func (f *InventoryProvider) ManifestReader(reader io.Reader, args []string) (manifestreader.ManifestReader, error) {
	// Fetch the namespace from the configloader. The source of this
	// either the namespace flag or the context. If the namespace is provided
	// with the flag, enforceNamespace will be true. In this case, it is
	// an error if any of the resources in the package has a different
	// namespace set.
	namespace, enforceNamespace, err := f.factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return nil, err
	}

	mapper, err := f.factory.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	readerOptions := manifestreader.ReaderOptions{
		Mapper:           mapper,
		Namespace:        namespace,
		EnforceNamespace: enforceNamespace,
	}

	var mReader manifestreader.ManifestReader
	if len(args) == 0 {
		mReader = &manifestreader.StreamManifestReader{
			ReaderName:    "stdin",
			Reader:        reader,
			ReaderOptions: readerOptions,
		}
	} else {
		mReader = &manifestreader.PathManifestReader{
			Path:          args[0],
			ReaderOptions: readerOptions,
		}
	}
	return mReader, nil
}

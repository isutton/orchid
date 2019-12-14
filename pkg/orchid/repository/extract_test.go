package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isutton/orchid/pkg/orchid/orm"
	"github.com/isutton/orchid/test/mocks"
)

func TestExtract_extract(t *testing.T) {
	cr, err := mocks.UnstructuredCRMock()
	require.NoError(t, err)

	data, err := extract(cr.Object, orm.JSTypeString, []string{"spec", "simple"})
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, "11", data)
}

func TestExtract_extractCRDOpenAPIV3Schema(t *testing.T) {
	crd, err := mocks.UnstructuredCRDMock()
	require.NoError(t, err)

	openAPIV3Schema, err := extractCRDOpenAPIV3Schema(crd.Object)
	require.NoError(t, err)
	assert.NotNil(t, openAPIV3Schema)
}

func TestExtract_extractCRGVKFromCRD(t *testing.T) {
	cr, err := mocks.UnstructuredCRDMock()
	require.NoError(t, err)

	gvk, err := extractCRGVKFromCRD(cr.Object)
	require.NoError(t, err)
	assert.NotNil(t, gvk)

	assert.Equal(t, "mock", gvk.Group)
	assert.Equal(t, "v1", gvk.Version)
	assert.Equal(t, "Custom", gvk.Kind)
}
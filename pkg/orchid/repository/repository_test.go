package repository

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/klogr"

	"github.com/isutton/orchid/pkg/orchid/config"
	jsc "github.com/isutton/orchid/pkg/orchid/jsonschema"
	"github.com/isutton/orchid/pkg/orchid/orm"
	"github.com/isutton/orchid/test/mocks"
)

func buildTestRepository(t *testing.T) (logr.Logger, *Repository) {
	logger := klogr.New().WithName("test")
	config := &config.Config{Username: "postgres", Password: "1", Options: "sslmode=disable"}

	r := NewRepository(logger, config)
	require.NotNil(t, r)
	return logger, r
}

func TestRepository_decompose(t *testing.T) {
	logger, repo := buildTestRepository(t)

	schemaName := "orchid"
	s := orm.NewSchema(logger, schemaName)

	openAPIV3Schema := jsc.ExtV1CRDOpenAPIV3Schema()
	err := s.Generate(&openAPIV3Schema)
	require.NoError(t, err)

	u, err := mocks.UnstructuredCRDMock()
	require.NoError(t, err)

	matrix, err := repo.decompose(s, u)
	require.NoError(t, err)
	require.NotNil(t, matrix)

	t.Logf("matrix='%#v'", matrix)

	table, err := s.GetTable(schemaName)
	require.NoError(t, err)

	// looking for column position in order to acquire data from matrix
	columnPosition := -1
	for position, column := range table.ColumNames() {
		if column == orm.XEmbeddedResource {
			columnPosition = position
			break
		}
	}
	require.NotEqual(t, -1, columnPosition)

	row, found := matrix[schemaName]
	require.True(t, found)
	require.Len(t, row, 1)

	jsonData, ok := row[0][columnPosition].(string)
	require.True(t, ok)
	assert.NotNil(t, jsonData)

	t.Logf("json='%s'", jsonData)

	// making sure json data also has a new line appended
	jsonData = fmt.Sprintf("%s\n", jsonData)

	// comparing json data with original object
	bytes, err := u.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, jsonData, string(bytes))
}

func TestRepository_New(t *testing.T) {
	_, repo := buildTestRepository(t)

	t.Run("Bootstrap", func(t *testing.T) {
		err := repo.Bootstrap()
		require.NoError(t, err)
	})

	t.Run("Create-CRD", func(t *testing.T) {
		crd, err := mocks.UnstructuredCRDMock()
		require.NoError(t, err)

		err = repo.Create(crd)
		require.NoError(t, err)
	})

	cr, err := mocks.UnstructuredCRMock()
	t.Logf("cr='%#v'", cr)
	require.NoError(t, err)

	gvk := cr.GetObjectKind().GroupVersionKind()

	o, s, err := repo.factory(DefaultNamespace, gvk)
	require.NoError(t, err)

	table, err := s.GetTable(s.Name)
	require.NoError(t, err)
	// cleaning up before other tests
	_, err = o.DB.Exec(fmt.Sprintf("truncate table %s cascade", table.Name))
	require.NoError(t, err)

	t.Run("Create-CR", func(t *testing.T) {
		err = repo.Create(cr)
		require.NoError(t, err)
	})

	t.Run("Read-CR", func(t *testing.T) {
		namespacedName := types.NamespacedName{
			Namespace: cr.GetNamespace(),
			Name:      cr.GetName(),
		}
		u, err := repo.Read(gvk, namespacedName)
		require.NoError(t, err)

		t.Logf("u='%#v'", u)

		t.Log("## comparing original and obtained unstructured")
		diff := deep.Equal(cr, u)
		for _, entry := range diff {
			t.Log(entry)
		}
	})

	t.Run("List-CR", func(t *testing.T) {
		err = repo.Create(cr)
		require.NoError(t, err)

		options := metav1.ListOptions{LabelSelector: "label=label"}
		list, err := repo.List("orchid", gvk, options)
		require.NoError(t, err)

		t.Logf("List size '%d'", len(list.Items))
		for _, item := range list.Items {
			t.Logf("item='%#v'", item.Object)
		}

		assert.Len(t, list.Items, 2)
	})
}

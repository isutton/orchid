package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumn_New(t *testing.T) {
	t.Run("NewColumn", func(t *testing.T) {
		column, err := NewColumn("test", "integer", "int32", true)
		assert.NoError(t, err)
		assert.NotEmpty(t, column.Type)
		assert.NotEmpty(t, column.String())
		assert.Contains(t, column.String(), "integer")
		assert.Contains(t, column.String(), "not null")
	})

	t.Run("NewColumnArray", func(t *testing.T) {
		var max int64 = 10
		column, err := NewColumnArray("test", "integer", "int32", &max, false)
		assert.NoError(t, err)
		assert.NotEmpty(t, column.Type)
		assert.NotEmpty(t, column.String())
		assert.Contains(t, column.String(), "integer[10]")
	})
}

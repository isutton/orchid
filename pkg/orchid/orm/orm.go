package orm

import (
	"database/sql"
	"fmt"

	"github.com/go-logr/logr"
	_ "github.com/lib/pq"
	"k8s.io/apimachinery/pkg/types"
)

// ORM represents the data abastraction layer.
type ORM struct {
	logger  logr.Logger // logger instance
	connStr string      // database adapter connection string
	DB      *sql.DB     // database adapter instance
}

type List []interface{}
type MappedList map[string]List
type MappedMatrix map[string][]List

type Entry map[string]interface{}
type EntryMap map[string]Entry
type MappedEntries map[string][]Entry

// CreateSchemaTables create tables for a schema.
func (o *ORM) CreateSchemaTables(schema *Schema) error {
	for _, statement := range CreateTablesStatement(schema) {
		_, err := o.DB.Query(statement)
		if err != nil {
			return err
		}
	}
	return nil
}

// Connect with the database, instantiate the connection.
func (o *ORM) Connect() error {
	var err error
	o.DB, err = sql.Open("postgres", o.connStr)
	return err
}

// interpolate table columne's argument with cached primary-keys, in order to complete the desired
// amount of columns with foreign-keys.
func (o *ORM) interpolate(table *Table, arguments List, cachedIDs map[string]int64) (List, error) {
	argumentWithFK := make(List, 0)
	pos := 0
	for _, column := range table.Columns {
		if table.IsPrimaryKey(column.Name) {
			continue
		}
		if targetFKTable := table.ForeignKeyTable(column.Name); targetFKTable != "" {
			foreingKeyID, found := cachedIDs[targetFKTable]
			if !found {
				return nil, fmt.Errorf("unable to find primary-key in cache '%#v'", cachedIDs)
			}
			argumentWithFK = append(argumentWithFK, foreingKeyID)
		} else {
			argumentWithFK = append(argumentWithFK, arguments[pos])
			pos += 1
		}
	}
	return argumentWithFK, nil
}

// resultMatrix build a matrix of results from sql.Rows.
func (o *ORM) resultMatrix(schema *Schema, rows *sql.Rows) (map[string]int, []List, error) {
	rowColumns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	// extracting row column names to create a map of name and column position
	columnIDs := make(map[string]int, len(schema.Tables))
	for i, name := range rowColumns {
		columnIDs[name] = i
	}

	matrix := make([]List, 0)
	// scanning row values to a single slice of slices
	for rows.Next() {
		columnValues := make(List, len(rowColumns))
		columnValuePointers := make(List, len(rowColumns))
		for i := range columnValues {
			columnValuePointers[i] = &columnValues[i]
		}
		// scanning results using pointers to populate columns slice
		if err = rows.Scan(columnValuePointers...); err != nil {
			return nil, nil, err
		}
		matrix = append(matrix, columnValues)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	return columnIDs, matrix, nil
}

// scanRows extract a matrix of results, and create a result-set object with it.
func (o *ORM) scanRows(schema *Schema, rows *sql.Rows) (*ResultSet, error) {
	// preparing row results into a single matrix of data, having back an map of column names and
	// their respective column number
	columnIDs, matrix, err := o.resultMatrix(schema, rows)
	if err != nil {
		return nil, err
	}

	return NewResultSet(schema, columnIDs, matrix)
}

// Create stores a given object in the database.
func (o *ORM) Create(schema *Schema, matrix MappedMatrix) error {
	rows := len(matrix)
	if rows == 0 {
		return fmt.Errorf("empty data informed")
	}
	logger := o.logger.WithValues("matrix-rows", rows, "schema", schema.Name)
	logger.Info("Executing create against informed schema.")

	statements := InsertStatement(schema)

	txn, err := o.DB.Begin()
	if err != nil {
		return err
	}

	tablePKCache := make(map[string]int64, len(statements))
	for i, table := range schema.Tables {
		statement := statements[i]
		arguments, found := matrix[table.Name]
		if !found {
			continue
		}
		logger = logger.WithValues(
			"statemnet", statement, "rows", len(arguments), "table", table.Name)

		// for each row found for that
		for _, argument := range arguments {
			logger.WithValues("columns", len(argument)).
				Info("Executing insert", "statement", statement)
			// in case the case of arguments for this table being less than expected, completing the
			// slice with foreign-key cached IDs
			if len(argument) == 0 {
				continue
			}
			// completing argument with foreign-keys values, cached from previous statements
			if len(argument) < len(table.Columns)-1 {
				if argument, err = o.interpolate(table, argument, tablePKCache); err != nil {
					return err
				}
			}
			// executing insert statement and capturing primary-key
			var primaryKeyValue int64
			if err = txn.QueryRow(statement, argument...).Scan(&primaryKeyValue); err != nil {
				return err
			}
			tablePKCache[table.Name] = primaryKeyValue
		}
	}

	return txn.Commit()
}

// Read namespaced-name from database, returned as a result-set instance.
func (o *ORM) Read(schema *Schema, namespacedName types.NamespacedName) (*ResultSet, error) {
	metadataTable, err := schema.GetTable(fmt.Sprintf("%s_metadata", schema.Name))
	if err != nil {
		return nil, err
	}
	whereNamespacedName := []string{
		fmt.Sprintf("%s.namespace", metadataTable.Hint),
		fmt.Sprintf("%s.name", metadataTable.Hint),
	}

	statement := SelectStatement(schema, whereNamespacedName)
	o.logger.WithValues("namespacedName", namespacedName).
		Info(statement)

	rows, err := o.DB.Query(statement, namespacedName.Namespace, namespacedName.Name)
	if err != nil {
		return nil, err
	}
	return o.scanRows(schema, rows)
}

// NewORM instantiate an ORM.
func NewORM(logger logr.Logger, connStr string) *ORM {
	return &ORM{logger: logger.WithName("orm"), connStr: connStr}
}

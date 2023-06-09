//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/sqlite"
)

var EventTypes = newEventTypesTable("", "event_types", "")

type eventTypesTable struct {
	sqlite.Table

	// Columns
	Name sqlite.ColumnString

	AllColumns     sqlite.ColumnList
	MutableColumns sqlite.ColumnList
}

type EventTypesTable struct {
	eventTypesTable

	EXCLUDED eventTypesTable
}

// AS creates new EventTypesTable with assigned alias
func (a EventTypesTable) AS(alias string) *EventTypesTable {
	return newEventTypesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new EventTypesTable with assigned schema name
func (a EventTypesTable) FromSchema(schemaName string) *EventTypesTable {
	return newEventTypesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new EventTypesTable with assigned table prefix
func (a EventTypesTable) WithPrefix(prefix string) *EventTypesTable {
	return newEventTypesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new EventTypesTable with assigned table suffix
func (a EventTypesTable) WithSuffix(suffix string) *EventTypesTable {
	return newEventTypesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newEventTypesTable(schemaName, tableName, alias string) *EventTypesTable {
	return &EventTypesTable{
		eventTypesTable: newEventTypesTableImpl(schemaName, tableName, alias),
		EXCLUDED:        newEventTypesTableImpl("", "excluded", ""),
	}
}

func newEventTypesTableImpl(schemaName, tableName, alias string) eventTypesTable {
	var (
		NameColumn     = sqlite.StringColumn("name")
		allColumns     = sqlite.ColumnList{NameColumn}
		mutableColumns = sqlite.ColumnList{}
	)

	return eventTypesTable{
		Table: sqlite.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		Name: NameColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}

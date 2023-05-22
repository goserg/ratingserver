//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var Roles = newRolesTable("public", "roles", "")

type rolesTable struct {
	postgres.Table

	// Columns
	ID postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type RolesTable struct {
	rolesTable

	EXCLUDED rolesTable
}

// AS creates new RolesTable with assigned alias
func (a RolesTable) AS(alias string) *RolesTable {
	return newRolesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new RolesTable with assigned schema name
func (a RolesTable) FromSchema(schemaName string) *RolesTable {
	return newRolesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new RolesTable with assigned table prefix
func (a RolesTable) WithPrefix(prefix string) *RolesTable {
	return newRolesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new RolesTable with assigned table suffix
func (a RolesTable) WithSuffix(suffix string) *RolesTable {
	return newRolesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newRolesTable(schemaName, tableName, alias string) *RolesTable {
	return &RolesTable{
		rolesTable: newRolesTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newRolesTableImpl("", "excluded", ""),
	}
}

func newRolesTableImpl(schemaName, tableName, alias string) rolesTable {
	var (
		IDColumn       = postgres.StringColumn("id")
		allColumns     = postgres.ColumnList{IDColumn}
		mutableColumns = postgres.ColumnList{}
	)

	return rolesTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID: IDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
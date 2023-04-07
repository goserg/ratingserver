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

var Users = newUsersTable("", "users", "")

type usersTable struct {
	sqlite.Table

	//Columns
	ID        sqlite.ColumnInteger
	FirstName sqlite.ColumnString
	Username  sqlite.ColumnString
	CreatedAt sqlite.ColumnTimestamp
	UpdatedAt sqlite.ColumnTimestamp

	AllColumns     sqlite.ColumnList
	MutableColumns sqlite.ColumnList
}

type UsersTable struct {
	usersTable

	EXCLUDED usersTable
}

// AS creates new UsersTable with assigned alias
func (a UsersTable) AS(alias string) *UsersTable {
	return newUsersTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new UsersTable with assigned schema name
func (a UsersTable) FromSchema(schemaName string) *UsersTable {
	return newUsersTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new UsersTable with assigned table prefix
func (a UsersTable) WithPrefix(prefix string) *UsersTable {
	return newUsersTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new UsersTable with assigned table suffix
func (a UsersTable) WithSuffix(suffix string) *UsersTable {
	return newUsersTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newUsersTable(schemaName, tableName, alias string) *UsersTable {
	return &UsersTable{
		usersTable: newUsersTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newUsersTableImpl("", "excluded", ""),
	}
}

func newUsersTableImpl(schemaName, tableName, alias string) usersTable {
	var (
		IDColumn        = sqlite.IntegerColumn("id")
		FirstNameColumn = sqlite.StringColumn("first_name")
		UsernameColumn  = sqlite.StringColumn("username")
		CreatedAtColumn = sqlite.TimestampColumn("created_at")
		UpdatedAtColumn = sqlite.TimestampColumn("updated_at")
		allColumns      = sqlite.ColumnList{IDColumn, FirstNameColumn, UsernameColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns  = sqlite.ColumnList{FirstNameColumn, UsernameColumn, CreatedAtColumn, UpdatedAtColumn}
	)

	return usersTable{
		Table: sqlite.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:        IDColumn,
		FirstName: FirstNameColumn,
		Username:  UsernameColumn,
		CreatedAt: CreatedAtColumn,
		UpdatedAt: UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}

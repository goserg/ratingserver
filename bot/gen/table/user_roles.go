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

var UserRoles = newUserRolesTable("", "user_roles", "")

type userRolesTable struct {
	sqlite.Table

	//Columns
	UserID sqlite.ColumnInteger
	RoleID sqlite.ColumnInteger

	AllColumns     sqlite.ColumnList
	MutableColumns sqlite.ColumnList
}

type UserRolesTable struct {
	userRolesTable

	EXCLUDED userRolesTable
}

// AS creates new UserRolesTable with assigned alias
func (a UserRolesTable) AS(alias string) *UserRolesTable {
	return newUserRolesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new UserRolesTable with assigned schema name
func (a UserRolesTable) FromSchema(schemaName string) *UserRolesTable {
	return newUserRolesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new UserRolesTable with assigned table prefix
func (a UserRolesTable) WithPrefix(prefix string) *UserRolesTable {
	return newUserRolesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new UserRolesTable with assigned table suffix
func (a UserRolesTable) WithSuffix(suffix string) *UserRolesTable {
	return newUserRolesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newUserRolesTable(schemaName, tableName, alias string) *UserRolesTable {
	return &UserRolesTable{
		userRolesTable: newUserRolesTableImpl(schemaName, tableName, alias),
		EXCLUDED:       newUserRolesTableImpl("", "excluded", ""),
	}
}

func newUserRolesTableImpl(schemaName, tableName, alias string) userRolesTable {
	var (
		UserIDColumn   = sqlite.IntegerColumn("user_id")
		RoleIDColumn   = sqlite.IntegerColumn("role_id")
		allColumns     = sqlite.ColumnList{UserIDColumn, RoleIDColumn}
		mutableColumns = sqlite.ColumnList{}
	)

	return userRolesTable{
		Table: sqlite.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		UserID: UserIDColumn,
		RoleID: RoleIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}

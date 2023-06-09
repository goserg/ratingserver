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

var Matches = newMatchesTable("", "matches", "")

type matchesTable struct {
	sqlite.Table

	// Columns
	ID        sqlite.ColumnInteger
	PlayerA   sqlite.ColumnString
	PlayerB   sqlite.ColumnString
	Winner    sqlite.ColumnString
	CreatedAt sqlite.ColumnTimestamp

	AllColumns     sqlite.ColumnList
	MutableColumns sqlite.ColumnList
}

type MatchesTable struct {
	matchesTable

	EXCLUDED matchesTable
}

// AS creates new MatchesTable with assigned alias
func (a MatchesTable) AS(alias string) *MatchesTable {
	return newMatchesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new MatchesTable with assigned schema name
func (a MatchesTable) FromSchema(schemaName string) *MatchesTable {
	return newMatchesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new MatchesTable with assigned table prefix
func (a MatchesTable) WithPrefix(prefix string) *MatchesTable {
	return newMatchesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new MatchesTable with assigned table suffix
func (a MatchesTable) WithSuffix(suffix string) *MatchesTable {
	return newMatchesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newMatchesTable(schemaName, tableName, alias string) *MatchesTable {
	return &MatchesTable{
		matchesTable: newMatchesTableImpl(schemaName, tableName, alias),
		EXCLUDED:     newMatchesTableImpl("", "excluded", ""),
	}
}

func newMatchesTableImpl(schemaName, tableName, alias string) matchesTable {
	var (
		IDColumn        = sqlite.IntegerColumn("id")
		PlayerAColumn   = sqlite.StringColumn("player_a")
		PlayerBColumn   = sqlite.StringColumn("player_b")
		WinnerColumn    = sqlite.StringColumn("winner")
		CreatedAtColumn = sqlite.TimestampColumn("created_at")
		allColumns      = sqlite.ColumnList{IDColumn, PlayerAColumn, PlayerBColumn, WinnerColumn, CreatedAtColumn}
		mutableColumns  = sqlite.ColumnList{PlayerAColumn, PlayerBColumn, WinnerColumn, CreatedAtColumn}
	)

	return matchesTable{
		Table: sqlite.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:        IDColumn,
		PlayerA:   PlayerAColumn,
		PlayerB:   PlayerBColumn,
		Winner:    WinnerColumn,
		CreatedAt: CreatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}

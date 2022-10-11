package sqlite3

import (
	"context"
	"os"
	"testing"

	"github.com/go-rel/rel"
	"github.com/go-rel/sql"
	"github.com/go-rel/sql/specs"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var ctx = context.TODO()

func dsn() string {
	if os.Getenv("SQLITE3_DATABASE") != "" {
		return os.Getenv("SQLITE3_DATABASE") + "?_foreign_keys=1&_loc=Local"
	}

	return "./rel_test.db?_foreign_keys=1&_loc=Local"
}

func TestAdapter_specs(t *testing.T) {
	adapter, err := Open(dsn())
	assert.Nil(t, err)
	defer adapter.Close()

	repo := rel.New(adapter)

	// Prepare tables
	teardown := specs.Setup(repo)
	defer teardown()

	// Migration Specs
	specs.Migrate(specs.SkipDropColumn)

	// Query Specs
	specs.Query(t, repo)
	specs.QueryJoin(t, repo)
	specs.QueryJoinAssoc(t, repo)
	specs.QueryWhereSubQuery(t, repo, specs.SkipAllAndAnyKeyword)
	specs.QueryNotFound(t, repo)

	// Preload specs
	specs.PreloadHasMany(t, repo)
	specs.PreloadHasManyWithQuery(t, repo)
	specs.PreloadHasManySlice(t, repo)
	specs.PreloadHasOne(t, repo)
	specs.PreloadHasOneWithQuery(t, repo)
	specs.PreloadHasOneSlice(t, repo)
	specs.PreloadBelongsTo(t, repo)
	specs.PreloadBelongsToWithQuery(t, repo)
	specs.PreloadBelongsToSlice(t, repo)

	// Aggregate Specs
	specs.Aggregate(t, repo)

	// Insert Specs
	specs.Insert(t, repo)
	specs.InsertHasMany(t, repo)
	specs.InsertHasOne(t, repo)
	specs.InsertBelongsTo(t, repo)
	specs.Inserts(t, repo)
	specs.InsertAll(t, repo)
	specs.InsertOnConflictIgnore(t, repo)
	specs.InsertOnConflictReplace(t, repo)
	specs.InsertAllOnConflictIgnore(t, repo)
	specs.InsertAllOnConflictReplace(t, repo)
	// specs.InsertAllPartialCustomPrimary(t, repo) - not supported

	// Update Specs
	specs.Update(t, repo)
	specs.UpdateNotFound(t, repo)
	specs.UpdateHasManyInsert(t, repo)
	specs.UpdateHasManyUpdate(t, repo)
	specs.UpdateHasManyReplace(t, repo)
	specs.UpdateHasOneInsert(t, repo)
	specs.UpdateHasOneUpdate(t, repo)
	specs.UpdateBelongsToInsert(t, repo)
	specs.UpdateBelongsToUpdate(t, repo)
	specs.UpdateAtomic(t, repo)
	specs.Updates(t, repo)
	specs.UpdateAny(t, repo)

	// Delete specs
	specs.Delete(t, repo)
	specs.DeleteBelongsTo(t, repo)
	specs.DeleteHasOne(t, repo)
	specs.DeleteHasMany(t, repo)
	specs.DeleteAll(t, repo)
	specs.DeleteAny(t, repo)

	// Constraint specs
	// - foreign key constraint is not supported because of lack of information in the error message.
	specs.UniqueConstraintOnInsert(t, repo)
	specs.UniqueConstraintOnUpdate(t, repo)
	specs.CheckConstraintOnInsert(t, repo)
	specs.CheckConstraintOnUpdate(t, repo)
}

func TestAdapter_Transaction_commitError(t *testing.T) {
	adapter, err := Open(dsn())
	assert.Nil(t, err)
	defer adapter.Close()

	assert.NotNil(t, adapter.Commit(ctx))
}

func TestAdapter_Transaction_rollbackError(t *testing.T) {
	adapter, err := Open(dsn())
	assert.Nil(t, err)
	defer adapter.Close()

	assert.NotNil(t, adapter.Rollback(ctx))
}

func TestAdapter_Exec_error(t *testing.T) {
	adapter, err := Open(dsn())
	assert.Nil(t, err)
	defer adapter.Close()

	_, _, err = adapter.Exec(ctx, "error", nil)
	assert.NotNil(t, err)
}

func TestAdapter_TableBuilder(t *testing.T) {
	adapter, err := Open(dsn())
	assert.Nil(t, err)
	defer adapter.Close()

	tests := []struct {
		result string
		table  rel.Table
	}{
		{
			result: `ALTER TABLE "columns" ADD COLUMN "verified" BOOL;ALTER TABLE "columns" RENAME COLUMN "string" TO "name";ALTER TABLE "columns" ;ALTER TABLE "columns" DROP COLUMN "blob";`,
			table: rel.Table{
				Op:   rel.SchemaAlter,
				Name: "columns",
				Definitions: []rel.TableDefinition{
					rel.Column{Name: "verified", Type: rel.Bool, Op: rel.SchemaCreate},
					rel.Column{Name: "string", Rename: "name", Op: rel.SchemaRename},
					rel.Column{Name: "bool", Type: rel.Int, Op: rel.SchemaAlter},
					rel.Column{Name: "blob", Op: rel.SchemaDrop},

					// unsupported and will be skipped
					rel.Key{Op: rel.SchemaCreate, Columns: []string{"user_id"}, Type: rel.ForeignKey, Reference: rel.ForeignKeyReference{Table: "products", Columns: []string{"id", "name"}}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.result, func(t *testing.T) {
			assert.Equal(t, test.result, adapter.(*sql.SQL).TableBuilder.Build(test.table))
		})
	}
}

package history

import (
	"context"
	"database/sql"
	"testing"

	"github.com/sanjayhashcash/go/services/aurora/internal/test"
)

func TestTryStateVerificationLock(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	test.ResetAuroraDB(t, tt.AuroraDB)
	q := &Q{tt.AuroraSession()}
	otherQ := &Q{q.Clone()}

	_, err := q.TryStateVerificationLock(context.Background())
	tt.Assert.EqualError(err, "cannot be called outside of a transaction")

	tt.Assert.NoError(q.BeginTx(tt.Ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	}))
	ok, err := q.TryStateVerificationLock(context.Background())
	tt.Assert.NoError(err)
	tt.Assert.True(ok)

	// lock is already held by q so we will not succeed
	tt.Assert.NoError(otherQ.BeginTx(tt.Ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	}))
	ok, err = otherQ.TryStateVerificationLock(context.Background())
	tt.Assert.NoError(err)
	tt.Assert.False(ok)

	// when q is rolled back that releases the lock
	tt.Assert.NoError(q.Rollback())

	// now otherQ is able to acquire the lock
	ok, err = otherQ.TryStateVerificationLock(context.Background())
	tt.Assert.NoError(err)
	tt.Assert.True(ok)

	tt.Assert.NoError(otherQ.Rollback())
}

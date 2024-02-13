package history

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sanjayhashcash/go/services/aurora/internal/test"
	"github.com/sanjayhashcash/go/xdr"
)

func TestLiquidityPoolLoader(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	test.ResetAuroraDB(t, tt.AuroraDB)
	session := tt.AuroraSession()

	var ids []string
	for i := 0; i < 100; i++ {
		poolID := xdr.PoolId{byte(i)}
		id, err := xdr.MarshalHex(poolID)
		tt.Assert.NoError(err)
		ids = append(ids, id)
	}

	loader := NewLiquidityPoolLoader()
	for _, id := range ids {
		future := loader.GetFuture(id)
		_, err := future.Value()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `invalid liquidity pool loader state,`)
		duplicateFuture := loader.GetFuture(id)
		assert.Equal(t, future, duplicateFuture)
	}

	assert.NoError(t, loader.Exec(context.Background(), session))
	assert.Panics(t, func() {
		loader.GetFuture("not-present")
	})

	q := &Q{session}
	for _, id := range ids {
		internalID, err := loader.GetNow(id)
		assert.NoError(t, err)
		lp, err := q.LiquidityPoolByID(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, lp.PoolID, id)
		assert.Equal(t, lp.InternalID, internalID)
	}

	_, err := loader.GetNow("not present")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `was not found`)
}

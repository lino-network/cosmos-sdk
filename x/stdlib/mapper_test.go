package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dbm "github.com/tendermint/tmlibs/db"

	abci "github.com/tendermint/abci/types"

	wire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type S struct {
	I int64
	B bool
}

func defaultComponents(key sdk.StoreKey) (sdk.Context, *wire.Codec) {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, nil)
	cdc := wire.NewCodec()
	cdc.RegisterConcrete(S{}, "S", nil)
	return ctx, cdc
}

func TestListMapper(t *testing.T) {
	key := sdk.NewKVStoreKey("list")
	ctx, cdc := defaultComponents(key)
	lm := NewListMapper(S{}, cdc, key)

	val := S{1, true}
	lm.Push(ctx, val)
	assert.Equal(t, int64(1), lm.Len(ctx))
	assert.Equal(t, val, *lm.Get(ctx, int64(0)).(*S))

	val = S{2, false}
	lm.Set(ctx, 0, val)
	assert.Equal(t, val, *lm.Get(ctx, int64(0)).(*S))
}

/*
func TestQueueMapper(t *testing.T) {
	key := sdk.NewKVStoreKey("queue")
	ctx, cdc := defaultComponents(key)
	qm := NewQueueMapper(S{}, cdc, key)

	val := S{1, true}
	qm.Push(ctx, val)
	assert.Equal(t, val, *qm.Peek(ctx).(*S))

	qm.Pop(ctx)
	assert.Equal(t, nil, *qm.Peek(ctx).(*S))
}*/

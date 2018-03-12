package stdlib

import (
	"errors"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wire "github.com/tendermint/go-wire"
)

type ListMapper interface { // Solidity list like structure
	Len(sdk.Context) int64
	Get(sdk.Context, int64) interface{}
	Set(sdk.Context, int64, interface{})
	Push(sdk.Context, interface{})
	Iterate(sdk.Context, func(sdk.Context, int64, interface{}))
}

type listMapper struct {
	key sdk.StoreKey
	cdc *wire.Codec
	lk  []byte
	vt  reflect.Type
}

func NewListMapper(value interface{}, cdc *wire.Codec, key sdk.StoreKey) ListMapper {
	rt := reflect.TypeOf(value)
	if rt.Kind() == reflect.Ptr {
		// TODO
		panic(errors.New(""))
	}
	lk, err := cdc.MarshalBinary(int64(-1))
	if err != nil {
		panic(err)
	}
	return listMapper{
		key: key,
		cdc: cdc,
		lk:  lk,
		vt:  rt,
	}
}

func (lm listMapper) Len(ctx sdk.Context) int64 {
	store := ctx.KVStore(lm.key)
	bz := store.Get(lm.lk)
	if bz == nil {
		zero, err := lm.cdc.MarshalBinary(0)
		if err != nil {
			panic(err)
		}
		store.Set(lm.lk, zero)
		return 0
	}
	var res int64
	if err := lm.cdc.UnmarshalBinary(bz, &res); err != nil {
		panic(err)
	}
	return res
}

func (lm listMapper) Get(ctx sdk.Context, index int64) interface{} {
	if index < 0 {
		panic(errors.New(""))
	}
	store := ctx.KVStore(lm.key)
	bz := store.Get(marshalInt64(lm.cdc, index))
	res := reflect.New(lm.vt).Interface()
	if err := lm.cdc.UnmarshalBinary(bz, res); err != nil {
		panic(err)
	}
	return res
}

func (lm listMapper) Set(ctx sdk.Context, index int64, value interface{}) {
	if index < 0 {
		panic(errors.New(""))
	}
	store := ctx.KVStore(lm.key)
	bz, err := lm.cdc.MarshalBinary(value)
	if err != nil {
		panic(err)
	}
	store.Set(marshalInt64(lm.cdc, index), bz)
}

func (lm listMapper) Push(ctx sdk.Context, value interface{}) {
	length := lm.Len(ctx)
	lm.Set(ctx, length, value)

	store := ctx.KVStore(lm.key)
	store.Set(lm.lk, marshalInt64(lm.cdc, length+1))
}

func (lm listMapper) Iterate(ctx sdk.Context, fn func(sdk.Context, int64, interface{})) {
	length := lm.Len(ctx)
	for i := int64(0); i < length; i++ {
		v := lm.Get(ctx, i)
		fn(ctx, i, v)
	}
}

type QueueMapper interface {
	Push(sdk.Context, interface{})
	Peek(sdk.Context) interface{}
	Pop(sdk.Context)
	Iterate(sdk.Context, func(sdk.Context, interface{}))
}

type queueMapper struct {
	key sdk.StoreKey
	cdc *wire.Codec
	ik  []byte
	vt  reflect.Type
}

func NewQueueMapper(value interface{}, cdc *wire.Codec, key sdk.StoreKey) QueueMapper {
	rt := reflect.TypeOf(value)
	if rt.Kind() == reflect.Ptr {
		// TODO
		panic(errors.New(""))
	}
	ik, err := cdc.MarshalBinary(int64(-1))
	if err != nil {
		panic(err)
	}
	return queueMapper{
		key: key,
		cdc: cdc,
		ik:  ik,
		vt:  rt,
	}
}

type queueInfo struct {
	// begin <= elems < end
	begin int64
	end   int64
}

func (info queueInfo) validateBasic() error {
	if info.end < info.begin || info.begin < 0 || info.end < 0 {
		return errors.New("")
	}
	return nil
}

func (info queueInfo) isEmpty() bool {
	return info.begin == info.end
}

func (qm queueMapper) getQueueInfo(store sdk.KVStore) queueInfo {
	bz := store.Get(qm.ik)
	if bz == nil {
		store.Set(qm.ik, marshalQueueInfo(qm.cdc, queueInfo{0, 0}))
		return queueInfo{0, 0}
	}
	var info queueInfo
	if err := qm.cdc.UnmarshalBinary(bz, &info); err != nil {
		panic(err)
	}
	if err := info.validateBasic(); err != nil {
		panic(err)
	}
	return info
}

func (qm queueMapper) setQueueInfo(store sdk.KVStore, info queueInfo) {
	bz, err := qm.cdc.MarshalBinary(info)
	if err != nil {
		panic(err)
	}
	store.Set(qm.ik, bz)
}

func (qm queueMapper) Push(ctx sdk.Context, value interface{}) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)

	bz, err := qm.cdc.MarshalBinary(value)
	if err != nil {
		panic(err)
	}
	store.Set(marshalInt64(qm.cdc, info.end), bz)

	info.end++
	qm.setQueueInfo(store, info)
}

func (qm queueMapper) Peek(ctx sdk.Context) interface{} {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)
	bz := store.Get(marshalInt64(qm.cdc, info.begin))
	res := reflect.New(qm.vt).Interface()
	if err := qm.cdc.UnmarshalBinary(bz, res); err != nil {
		panic(err)
	}
	return res
}

func (qm queueMapper) Pop(ctx sdk.Context) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)
	store.Delete(marshalInt64(qm.cdc, info.begin))
	info.begin++
	qm.setQueueInfo(store, info)
}

func (qm queueMapper) Iterate(ctx sdk.Context, fn func(sdk.Context, interface{})) {
	store := ctx.KVStore(qm.key)
	info := qm.getQueueInfo(store)

	for i := info.begin; i < info.end; i++ {
		bz := store.Get(marshalInt64(qm.cdc, i))
		value := reflect.New(qm.vt).Interface()
		if err := qm.cdc.UnmarshalBinary(bz, value); err != nil {
			panic(err)
		}
		fn(ctx, value)
	}

	info.begin = info.end
	qm.setQueueInfo(store, info)
}

func marshalQueueInfo(cdc *wire.Codec, info queueInfo) []byte {
	bz, err := cdc.MarshalBinary(info)
	if err != nil {
		panic(err)
	}
	return bz
}

func marshalInt64(cdc *wire.Codec, i int64) []byte {
	bz, err := cdc.MarshalBinary(i)
	if err != nil {
		panic(err)
	}
	return bz
}

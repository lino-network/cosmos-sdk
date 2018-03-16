package commands

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/builder"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/client/keys"

	"github.com/cosmos/cosmos-sdk/x/bank"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
)

// SendTxCommand will create a send tx and sign it with the given key
func SendTxCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := commander{cdc}
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE:  cmdr.sendTxCmd,
	}
	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	return cmd
}

type commander struct {
	cdc *wire.Codec
}

func (c commander) sendTxCmd(cmd *cobra.Command, args []string) error {

	// get the from address
	from, err := builder.GetFromAddress()
	if err != nil {
		return err
	}

	// build send msg
	msg, err := buildMsg(from)
	if err != nil {
		return err
	}
	chainID := "test-chain-P6aQsS"
	sequence := int64(viper.GetInt(client.FlagSequence))

	signMsg := sdk.StdSignMsg{
		ChainID:   chainID,
		Sequences: []int64{sequence},
		Msg:       msg,
	}

	// build and sign the transaction, then broadcast to Tendermint
	res, err := builder.SignBuildBroadcast(signMsg, c.cdc)
	if err != nil {
		return err
	}

	fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
	return nil
}

func buildMsg(from sdk.Address) (sdk.Msg, error) {
	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	name := "yukaitu"

	info, err := keybase.Get(name)
	if err != nil {
		return nil, nil
	}

	name1 := viper.GetString(client.FlagName)
	info2, err := keybase.Get(name1)
	if err != nil {
		return nil, nil
	}

	fmt.Println(reflect.DeepEqual(info, info2))

	fmt.Println(reflect.DeepEqual(info.Address, info2.Address))
	// parse coins

	info3, err := keybase.Get(name)
	if err != nil {
		return nil, nil
	}
	fmt.Println(info.PubKey.Address(), info3.PubKey.Address())
	fmt.Println(reflect.DeepEqual(info.PubKey.Address(), info3.PubKey.Address()))
	fmt.Println(info.PubKey.Address() == info3.PubKey.Address())
	// parse coins

	amount := viper.GetString(flagAmount)
	coins, err := sdk.ParseCoins(amount)
	if err != nil {
		return nil, err
	}

	// parse destination address
	dest := viper.GetString(flagTo)
	bz, err := hex.DecodeString(dest)
	if err != nil {
		return nil, err
	}
	to := sdk.Address(bz)

	input := bank.NewInput(info.PubKey.Address(), coins)
	output := bank.NewOutput(to, coins)
	msg := bank.NewSendMsg([]bank.Input{input}, []bank.Output{output})
	return msg, nil
}

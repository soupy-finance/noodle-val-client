package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/soupy-finance/noodle-val-client/pricegetter"
	"github.com/soupy-finance/noodle/x/oracle/types"
)

const AddressPrefix = "soupy"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	config := sdktypes.GetConfig()
	config.SetBech32PrefixForAccount(AddressPrefix, AddressPrefix+"pub")
	cosmos, err := cosmosclient.New(
		context.Background(),
		cosmosclient.WithAddressPrefix(AddressPrefix),
	)

	if err != nil {
		log.Fatal(err)
	}

	accountName := "alice"
	address, err := cosmos.Address(accountName)

	if err != nil {
		log.Fatal(err)
	}

	assets := getAssets()
	prices := map[string]string{}

	for _, asset := range assets {
		prices[asset] = "0"
	}

	go sendPrices(cosmos, address, accountName, prices)
	pricegetter.GetPrices()
}

func getAssets() []string {
	return []string{}
}

func sendPrices(
	cosmos cosmosclient.Client,
	address sdktypes.AccAddress,
	accountName string,
	prices map[string]string,
) {
	var i float64 = 1

	for {
		pricesJson, err := json.Marshal(prices)

		if err != nil {
			log.Fatal(err)
		}

		msg := &types.MsgUpdatePrices{
			Creator: address.String(),
			Data:    string(pricesJson),
		}
		_, err = cosmos.BroadcastTx(accountName, msg)
		fmt.Println("Price updated")

		if err != nil {
			log.Fatal(err)
		}

		// Testing
		prices["eth"] = fmt.Sprintf("%f", i)
		i += 1
	}
}

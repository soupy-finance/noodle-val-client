package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/soupy-finance/noodle-val-client/pricesgetter"
	"github.com/soupy-finance/noodle/x/oracle/types"
	oracletypes "github.com/soupy-finance/noodle/x/oracle/types"
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

	assets := getAssets(cosmos)
	prices := map[string]string{}

	for _, asset := range assets {
		prices[asset] = "0"
	}

	wsClosed := make(chan bool)
	go pricesgetter.GetPrices(assets, prices, wsClosed, interrupt)
	sendPrices(cosmos, address, accountName, prices, wsClosed, interrupt)
}

func getAssets(cosmos cosmosclient.Client) []string {
	queryClient := oracletypes.NewQueryClient(cosmos.Context())

	res, err := queryClient.Params(context.Background(), &oracletypes.QueryParamsRequest{})

	if err != nil {
		log.Fatal(err)
	}

	assets := []string{}
	json.Unmarshal([]byte(res.Params.Assets), &assets)
	return assets
}

func sendPrices(
	cosmos cosmosclient.Client,
	address sdktypes.AccAddress,
	accountName string,
	prices map[string]string,
	wsClosed chan bool,
	interrupt chan os.Signal,
) {
	for {
		select {
		case <-interrupt:
			<-wsClosed
			return
		default:
			pricesJson, err := json.Marshal(prices)

			if err != nil {
				log.Fatal(err)
			}

			msg := &types.MsgUpdatePrices{
				Creator: address.String(),
				Data:    string(pricesJson),
			}
			_, err = cosmos.BroadcastTx(accountName, msg)

			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

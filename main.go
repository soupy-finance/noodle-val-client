package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/soupy-finance/noodle-val-client/faucetserver"
	"github.com/soupy-finance/noodle-val-client/pricesgetter"
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
	pricesMu := &sync.Mutex{}
	txMsgs := []sdktypes.Msg{}
	txMsgsMu := &sync.Mutex{}

	for _, asset := range assets {
		prices[asset] = "0"
	}

	getterStopped := make(chan bool)
	faucetStopped := make(chan bool)

	go pricesgetter.GetPrices(assets, prices, pricesMu, getterStopped, interrupt)
	go faucetserver.ListenAndServe(cosmos, address, accountName, &txMsgs, txMsgsMu, faucetStopped, interrupt)

	for {
		select {
		case <-interrupt:
			<-getterStopped
			<-faucetStopped
			return
		default:
			txMsgsMu.Lock()
			msg := createPricesMsg(cosmos, address, prices, pricesMu)
			txMsgs = append(txMsgs, msg)
			_, _ = cosmos.BroadcastTx(accountName, txMsgs...)
			txMsgs = []sdktypes.Msg{}
			txMsgsMu.Unlock()

			time.Sleep(100 * time.Millisecond)
		}
	}
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

func createPricesMsg(
	cosmos cosmosclient.Client,
	address sdktypes.AccAddress,
	prices map[string]string,
	pricesMu *sync.Mutex,
) *oracletypes.MsgUpdatePrices {
	pricesMu.Lock()
	pricesJson, err := json.Marshal(prices)
	pricesMu.Unlock()

	if err != nil {
		log.Fatal(err)
	}

	msg := &oracletypes.MsgUpdatePrices{
		Creator: address.String(),
		Data:    string(pricesJson),
	}
	return msg
}

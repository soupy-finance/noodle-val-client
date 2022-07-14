package faucetserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	bridgetypes "github.com/soupy-finance/noodle/x/bridge/types"
)

func ListenAndServe(
	cosmos cosmosclient.Client,
	address sdktypes.AccAddress,
	accountName string,
	txMsgs *[]sdktypes.Msg,
	txMsgsMu *sync.Mutex,
	faucetStopped chan bool,
	interrupt chan os.Signal,
) {
	defer close(faucetStopped)

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	mux.HandleFunc("/deposit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Deposit request")
		addr := r.URL.Query().Get("address")

		msg1 := &bridgetypes.MsgObserveDeposit{
			Creator:   address.String(),
			ChainId:   "ethereum",
			Depositor: addr,
			DepositId: "0",
			Quantity:  "10000",
			Asset:     "usdc",
		}
		msg2 := &bridgetypes.MsgObserveDeposit{
			Creator:   address.String(),
			ChainId:   "ethereum",
			Depositor: addr,
			DepositId: "0",
			Quantity:  "10",
			Asset:     "eth",
		}

		txMsgsMu.Lock()
		*txMsgs = append(*txMsgs, msg1, msg2)
		txMsgsMu.Unlock()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
	})

	go srv.ListenAndServe()
	fmt.Println("Faucet server started")
	<-interrupt
	srv.Shutdown(context.Background())
}

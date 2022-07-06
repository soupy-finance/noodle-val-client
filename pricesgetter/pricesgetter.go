package pricesgetter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type BinanceTickerMsg struct {
	Stream string
	Data   map[string]interface{}
}

var binanceUrl string = "wss://stream.binance.com:9443/stream?streams="

// var ftxUrl string = "wss://ftx.com/ws"
// var coinbaseUrl string = "wss://ws-feed.exchange.coinbase.com"

var binanceBroken chan bool

// var ftxBroken chan bool
// var coinbaseBroken chan bool

func GetPrices(assets []string, prices map[string]string, wsClosed chan bool, interrupt chan os.Signal) {
	defer close(wsClosed)

	binanceBroken = make(chan bool)
	// ftxBroken = make(chan bool)
	// coinbaseBroken = make(chan bool)

	go binanceConnect(assets, prices, interrupt)
	// go ftxConnect(assets, prices, interrupt)
	// go coinbaseConnect(assets, prices, interrupt)

	for {
		select {
		case <-interrupt:
			<-binanceBroken
			// <- ftxBroken
			// <- coinbaseBroken
			return
		case <-binanceBroken:
			go binanceConnect(assets, prices, interrupt)
			// case <-ftxBroken:
			// 	go ftxConnect(assets, prices, interrupt)
			// case <-coinbaseBroken:
			// 	go coinbaseConnect(assets, prices, interrupt)
		}
	}
}

func binanceConnect(assets []string, prices map[string]string, interrupt chan os.Signal) {
	streams := []string{}

	for _, asset := range assets {
		streams = append(streams, asset+"usdt@bookTicker")
	}

	query := strings.Join(streams, "/")
	conn, _, err := websocket.DefaultDialer.Dial(binanceUrl+query, nil)

	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	binanceHandler(conn, prices, interrupt)
}

func binanceHandler(conn *websocket.Conn, prices map[string]string, interrupt chan os.Signal) {
	defer close(binanceBroken)
	for {
		select {
		case <-interrupt:
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			conn.Close()
			return
		default:
			_, message, err := conn.ReadMessage()

			if err != nil {
				return
			}

			msgJson := BinanceTickerMsg{}
			err = json.Unmarshal(message, &msgJson)

			if err != nil {
				log.Fatal(err)
			}

			asset := strings.Replace(strings.ToLower(msgJson.Data["s"].(string)), "usdt", "", 1)
			bidPrice, err := strconv.ParseFloat(msgJson.Data["b"].(string), 64)

			if err != nil {
				log.Fatal(err)
			}

			askPrice, err := strconv.ParseFloat(msgJson.Data["a"].(string), 64)

			if err != nil {
				log.Fatal(err)
			}

			prices[asset] = fmt.Sprintf("%f", (bidPrice+askPrice)/2)
		}
	}
}

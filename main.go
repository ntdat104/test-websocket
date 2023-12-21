package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int      `json:"id"`
}

type KlineData struct {
	Stream string `json:"stream"`
	Data   struct {
		Event  string `json:"e"`
		Time   int64  `json:"E"`
		Symbol string `json:"s"`
		Kline  struct {
			OpenTime           int64  `json:"t"`
			CloseTime          int64  `json:"T"`
			Interval           string `json:"i"`
			FirstTradeID       int64  `json:"f"`
			LastTradeID        int64  `json:"L"`
			Open               string `json:"o"`
			Close              string `json:"c"`
			High               string `json:"h"`
			Low                string `json:"l"`
			Volume             string `json:"v"`
			NumberOfTrades     int64  `json:"n"`
			IsClosed           bool   `json:"x"`
			QuoteVolume        string `json:"q"`
			VolumeInAsset      string `json:"V"`
			QuoteVolumeInAsset string `json:"Q"`
			Ignore             string `json:"B"`
		} `json:"k"`
	} `json:"data"`
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: "stream.binance.com", Path: "/stream"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var kline KlineData
			errJson := json.Unmarshal(message, &kline)
			if errJson != nil {
				fmt.Println("Error:", err)
				return
			}

			if kline.Data.Symbol == "" {
				continue
			}

			// Convert UNIX timestamp to formatted time string
			formattedTime := time.Unix(kline.Data.Time/1000, 0).Format("15:04:05")

			// Print formatted output including Volume
			output := fmt.Sprintf("Time: %s | Symbol: %s | Open: %s | High: %s | Low: %s | Close: %s | Volume: %s",
				formattedTime,
				kline.Data.Symbol,
				kline.Data.Kline.Open,
				kline.Data.Kline.High,
				kline.Data.Kline.Low,
				kline.Data.Kline.Close,
				kline.Data.Kline.Volume,
			)

			fmt.Println(output)
		}
	}()

	// Creating the message
	message := WebSocketMessage{
		Method: "SUBSCRIBE",
		Params: []string{
			// "!miniTicker@arr@3000ms",
			// "btcusdt@aggTrade",
			// "btcusdt@depth",
			"btcusdt@kline_1s",
			"ethusdt@kline_1s",
			"bnbusdt@kline_1s",
		},
		ID: 1,
	}

	// Sending the message as JSON
	err = c.WriteJSON(message)
	if err != nil {
		log.Fatal("write:", err)
	}

	// Wait for interruption signal to exit
	select {
	case <-done:
	case <-interrupt:
		log.Println("interrupt")
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("write close:", err)
			return
		}
		select {
		case <-done:
		}
		return
	}
}

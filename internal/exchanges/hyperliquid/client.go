package hyperliquid

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"planner/internal/exchanges"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type HyperliquidClient struct {
	wsUrl       string
	conn        *websocket.Conn
	isConnected bool
	symbol      string
	priceChan   chan exchanges.PriceUpdate
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

func NewHyperliquidClient(symbol string) *HyperliquidClient {
	ctx, cancel := context.WithCancel(context.Background())

	return &HyperliquidClient{
		wsUrl:     "wss://api.hyperliquid.xyz/ws",
		symbol:    symbol,
		priceChan: make(chan exchanges.PriceUpdate, 10),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (c *HyperliquidClient) Subscribe() error {
	c.mu.RLock()
	if !c.isConnected || c.conn == nil {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	c.mu.RUnlock()

	subscription := map[string]any{
		"method": "subscribe",
		"subscription": map[string]string{
			"type": "trades", // here should be price feed, not trades, should thing about
			"coin": c.symbol,
		},
	}

	if err := c.conn.WriteJSON(subscription); err != nil {
		return err
	}

	log.Printf("Subscribed to %s on Hyperliquid", c.symbol)
	return nil
}

func (c *HyperliquidClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isConnected {
		return nil
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(c.wsUrl, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	c.isConnected = true
	go c.readMessages()

	log.Printf("Connected to Hyperliquid WebSocket for %s", c.symbol)

	return nil
}

func (c *HyperliquidClient) readMessages() {
	defer func() {
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}
			var response HyperliquidResponse
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}
			if priceData := c.extractPrice(response); priceData.Price > 0 {
				select {
				case c.priceChan <- priceData:
				case <-c.ctx.Done():
					return
				}
			}
		}
	}
}

func (c *HyperliquidClient) extractPrice(response HyperliquidResponse) exchanges.PriceUpdate {

	if len(response.Data) == 0 {
		return exchanges.PriceUpdate{}
	}

	price, err := strconv.ParseFloat(response.Data[0].Px, 64)
	if err != nil {
		log.Printf("Error parsing price: %v", err)
		return exchanges.PriceUpdate{}
	}

	return exchanges.PriceUpdate{
		Symbol:    "btc",
		Price:     price,
		Timestamp: response.Data[0].Time,
		Exchange:  "Hyperliquid",
	}
}

func (c *HyperliquidClient) PriceChannel() <-chan exchanges.PriceUpdate {
	return c.priceChan
}

func (c *HyperliquidClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return nil
	}

	c.cancel()

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}

	c.isConnected = false
	close(c.priceChan)

	log.Println("Closed Hyperliquid WebSocket connection")
	return nil
}

func (c *HyperliquidClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

func (c *HyperliquidClient) Name() string {
	return "Hyperliquid"
}

var ErrNotConnected = errors.New("not connected to exchange")

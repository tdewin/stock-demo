package main

type BuyTransaction struct {
	Request         float64
	Bought          float64
	Remaining       float64
	Product         string
	StockCalculated bool
	Price           float64
	Sum             float64
}

//for templates

type StockMessage struct {
	Stocks []Stock
	Admin  bool
}
type Stock struct {
	BuyID        int
	Product      string
	StockMessage string
	Stock        float64
	Unit         string
	Price        float64
}

type Message struct {
	MessageType string
	Message     string
	Pre         string
	Refresh     int
	Redirect    string
	LinkBack    string
	BuySum      float64
	BuyTable    *map[string]*BuyTransaction
}

func NewError(message string) *Message {
	return &Message{"danger", message, "Error", 12, "", "", 0, nil}
}

func NewMessage(message string) *Message {
	return &Message{"success", message, "Info", 0, "", "../", 0, nil}
}

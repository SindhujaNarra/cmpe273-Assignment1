package main

import (
	"fmt"
	//"net"
	//"net/rpc"
	"log"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
)

type Request struct {
	StockSymbolAndPercentage string
	Budget                   float64
}
type Response struct {
	TradeId        int
	Stocks         []string
	UnvestedAmount float64
}
type RequestPortfolio struct {
	TradeId int
}

type ResponsePortfolio struct {
	Stocks             []string
	CurrentMarketValue float64
	UnvestedAmount     float64
}

var tradeId int

type Obj struct {
	StorePf map[int](*PortfolioObj)
}
type PortfolioObj struct {
	Stocks         map[string](*ShareObj)
	UnvestedAmount float64
}
type ShareObj struct {
	PurchasedPrice float64
	noOfShares     float64
}

func main() {

	if len(os.Args) == 3 {
		callPart1()
	} else if len(os.Args) == 2 {
		callPart2()
	} else {
		fmt.Println("Usage: ", os.Args[0], "localhost:1237")
		log.Fatal(1)
	}
}

func callPart1() {

	var req Request
	var res Response

	req.StockSymbolAndPercentage = os.Args[1]
	req.Budget, _ = strconv.ParseFloat(os.Args[2], 64)

	client, err := jsonrpc.Dial("tcp", "127.0.0.1:1237")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	err = client.Call("Obj.Part1", req, &res)
	if err != nil {
		log.Fatal("error:", err)
	}

	//printing part1 response

	fmt.Println("\nResponse from server: ")
	fmt.Println("\nTradeId: ", res.TradeId)
	fmt.Println("\nStocks: ", res.Stocks)
	fmt.Println("\nUnvested amount: ", res.UnvestedAmount)

}

func callPart2() {

	// Portfolio

	var reqPortfolio RequestPortfolio
	var resPortfolio ResponsePortfolio

	tradeId, _ := strconv.ParseInt(os.Args[1], 10, 32)
	reqPortfolio.TradeId = int(tradeId)

	client, err := jsonrpc.Dial("tcp", "127.0.0.1:1237")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	err = client.Call("Obj.Part2", reqPortfolio, &resPortfolio)
	if err != nil {
		log.Fatal("error:", err)
	}

	fmt.Println("Stocks: ", resPortfolio.Stocks)
	fmt.Println("Current Market value: ", resPortfolio.CurrentMarketValue)
	fmt.Println("Unvested amount: ", resPortfolio.UnvestedAmount)

}

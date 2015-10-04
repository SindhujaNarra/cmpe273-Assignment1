package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"strings"
)

type Request struct {
	StockSymbolAndPercentage string
	Budget                   float64
}
type Response struct {
	TradeId        int
	Stocks         []string
	UnvestedAmount float64
	Count          int
}
type RequestPortfolio struct {
	TradeId int
}

type ResponsePortfolio struct {
	Stocks             []string
	CurrentMarketValue float64
	UnvestedAmount     float64
}

type Stock struct {
	List struct {
		Resources []struct {
			Resource struct {
				Fields struct {
					Name    string `json:"name"`
					Price   string `json:"price"`
					Symbol  string `json:"symbol"`
					Ts      string `json:"ts"`
					Type    string `json:"type"`
					UTCTime string `json:"utctime"`
					Volume  string `json:"volume"`
				} `json:"fields"`
			} `json:"resource"`
		} `json:"resources"`
	} `json:"list"`
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

//var resultStr []Response

func (t *Obj) Part1(req Request, res *Response) error {

	fmt.Println("\n\tInside Part1: Buy Shares")

	StockAndPercentage := strings.Split(req.StockSymbolAndPercentage, ",")

	tradeId++
	res.TradeId = tradeId

	if t.StorePf == nil { // --> storing the tradeid in map to retrieve portfolio

		t.StorePf = make(map[int](*PortfolioObj))

		t.StorePf[tradeId] = new(PortfolioObj)
		t.StorePf[tradeId].Stocks = make(map[string]*ShareObj)

	}

	for _, stockcnt := range StockAndPercentage {

		splitString := strings.Split(stockcnt, ":")
		StockName := splitString[0]
		str1 := splitString[1]
		str2 := strings.TrimSuffix(str1, "%")
		StockPer, _ := strconv.ParseFloat(str2, 64)

		amount := req.Budget * (StockPer / 100)

		fmt.Println("\n\t\tStockName: ", StockName)
		fmt.Println("\n\t\tStockPer: ", StockPer)
		fmt.Println("\n\t\tAmount: ", amount)

		// calling yahoo finance api
		url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json", StockName)
		urlRes, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer urlRes.Body.Close()

		body, err := ioutil.ReadAll(urlRes.Body) //--> reading response from yfa
		if err != nil {
			panic(err)
		}

		var stock Stock

		err = json.Unmarshal(body, &stock) // --> unmarshalling parameters
		if err != nil {
			panic(err)
		}

		fmt.Println(stock.List.Resources[0].Resource.Fields.Name)
		fmt.Println(stock.List.Resources[0].Resource.Fields.Symbol)
		fmt.Println(stock.List.Resources[0].Resource.Fields.Price)

		//stockPrice ,err := strconv.ParseFloat(stock.List.Resources[0].Resource.Fields.Price)

		stockPrice, _ := strconv.ParseFloat(stock.List.Resources[0].Resource.Fields.Price, 64)

		//stockPrice := callAPI(StockName)

		noOfShares := math.Floor(amount / stockPrice)

		res.UnvestedAmount += amount - (stockPrice * noOfShares)

		resultStr := StockName + ":" + strconv.FormatFloat(noOfShares, 'f', 0, 64) + ":$" + strconv.FormatFloat(stockPrice, 'f', 2, 64)

		res.Count++
		res.Stocks = append(res.Stocks, resultStr)

		// storing in map for checking Portfolio

		if _, ok := t.StorePf[tradeId]; !ok {

			pfObj := new(PortfolioObj)
			pfObj.Stocks = make(map[string]*ShareObj)
			t.StorePf[tradeId] = pfObj
		}
		if _, ok := t.StorePf[tradeId].Stocks[StockName]; !ok {

			shObj := new(ShareObj)
			shObj.PurchasedPrice = stockPrice
			shObj.noOfShares = noOfShares
			t.StorePf[tradeId].Stocks[StockName] = shObj

			fmt.Println("\n\t\t", shObj)
			fmt.Println("\n\t\t", t.StorePf[tradeId].Stocks[StockName])

		} else {

			total := (noOfShares * stockPrice) + (t.StorePf[tradeId].Stocks[StockName].noOfShares)*t.StorePf[tradeId].Stocks[StockName].PurchasedPrice

			t.StorePf[tradeId].Stocks[StockName].PurchasedPrice = total / (noOfShares + t.StorePf[tradeId].Stocks[StockName].noOfShares)

			t.StorePf[tradeId].Stocks[StockName].noOfShares += noOfShares
			fmt.Println("\nTotal: ", total)
		}

		//unvestedAmount := Budget - totalSpent
		unvestedAmt := res.UnvestedAmount
		t.StorePf[tradeId].UnvestedAmount += unvestedAmt

	}
	return nil

}

/***********************  PART 2 **************************************************************** */

func (t *Obj) Part2(reqPortfolio RequestPortfolio, res *ResponsePortfolio) error {

	fmt.Println("\n\tInside Part2: Check Portfolio")

	var tempRes string
	if objVal, ok := t.StorePf[tradeId]; ok {

		var currMarketVal float64
		for StockName, t := range objVal.Stocks {

			// calling yahoo finance api
			url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json", StockName)
			urlRes, err := http.Get(url)
			if err != nil {
				panic(err)
			}
			defer urlRes.Body.Close()

			body, err := ioutil.ReadAll(urlRes.Body) //--> reading response from yfa
			if err != nil {
				panic(err)
			}

			var stocks Stock

			err = json.Unmarshal(body, &stocks) // --> unmarshalling parameters
			if err != nil {
				panic(err)
			}

			fmt.Println(stocks.List.Resources[0].Resource.Fields.Name)
			fmt.Println(stocks.List.Resources[0].Resource.Fields.Symbol)
			fmt.Println(stocks.List.Resources[0].Resource.Fields.Price)

			stockPrice, _ := strconv.ParseFloat(stocks.List.Resources[0].Resource.Fields.Price, 64)

			//stockPrice := callAPI(StockName)

			if t.PurchasedPrice < stockPrice { // Profit or loss
				tempRes = "+$" + strconv.FormatFloat(stockPrice, 'f', 2, 64)
			} else if t.PurchasedPrice > stockPrice {
				tempRes = "-$" + strconv.FormatFloat(stockPrice, 'f', 2, 64)
			} else {
				tempRes = "$" + strconv.FormatFloat(stockPrice, 'f', 2, 64)
			}

			stock1 := StockName + ":" + strconv.FormatFloat(t.noOfShares, 'f', 0, 64) + ":" + tempRes
			fmt.Print("Stocks...", stock1)
			res.Stocks = append(res.Stocks, stock1)
			fmt.Print("Stocks...", res.Stocks)

			currMarketVal += t.noOfShares * stockPrice

		}
		res.UnvestedAmount = objVal.UnvestedAmount
		res.CurrentMarketValue = currMarketVal

	} else {
		fmt.Println("\n\n\tInvalid TradeId !!!")
		//return errors.New("Tradeid doesn't exist")
	}

	return nil

}

/*func callAPI(StockName string) float64 {

	url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json", StockName)
	urlRes, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer urlRes.Body.Close()

	body, err := ioutil.ReadAll(urlRes.Body) //--> reading response from yfa
	if err != nil {
		panic(err)
	}

	var stocks Stock

	err = json.Unmarshal(body, &stocks) // --> unmarshalling parameters
	if err != nil {
		panic(err)
	}

	fmt.Println(stocks.List.Resources[0].Resource.Fields.Name)
	fmt.Println(stocks.List.Resources[0].Resource.Fields.Symbol)
	fmt.Println(stocks.List.Resources[0].Resource.Fields.Price)

	stockPrice, _ := strconv.ParseFloat(stocks.List.Resources[0].Resource.Fields.Price, 64)

	return stockPrice

} */
func main() {

	cal := new(Obj)
	rpc.Register(cal)

	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1237")
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()

		if err != nil {
			continue
		}
		jsonrpc.ServeConn(conn)
		conn.Close()
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}

}

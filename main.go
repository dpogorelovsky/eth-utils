package main

import (
	"flag"
	"fmt"
	"info/client"
	"net/http"
	"strconv"
)

var hostFlag = flag.String("h", "http://localhost:9545", "host address of ethereum blockchain")
var localPortFlag = flag.String("p", "12345", "local port for the HTTP server")
var cl *client.Client

func main() {
	flag.Parse()

	cl = client.New(*hostFlag)
	fmt.Println("started eth client on ", *hostFlag)

	http.HandleFunc("/latestBlock", latestBlock)
	http.HandleFunc("/last10tx", last10tx)
	http.HandleFunc("/balance", balance)
	http.HandleFunc("/sendEth", sendEth)
	http.HandleFunc("/chainID", chainID)

	fmt.Printf("local server is running on http://localhost:%s\n", *localPortFlag)
	err := http.ListenAndServe(":"+*localPortFlag, nil)
	if err != nil {
		fmt.Println("http server error: ", err.Error())
	}
	fmt.Println("stopper sever")
}

func latestBlock(w http.ResponseWriter, req *http.Request) {
	resp, err := cl.LatestBlock()
	if err != nil {
		resp = "Error: " + err.Error()
	}
	fmt.Fprintf(w, "%s\n", resp)
}

func last10tx(w http.ResponseWriter, req *http.Request) {
	resp, err := cl.Last10Tx()
	if err != nil {
		resp = "Error: " + err.Error()
	}
	fmt.Fprintf(w, "%s\n", resp)
}

func balance(w http.ResponseWriter, req *http.Request) {
	acc := req.URL.Query().Get("acc")
	if acc == "" {
		fmt.Fprintf(w, "%s\n", "Error: missing acc parameter")
		return
	}
	resp, err := cl.Balance(acc)
	if err != nil {
		resp = "Error: " + err.Error()
	}
	fmt.Fprintf(w, "%s\n", resp)
}

func sendEth(w http.ResponseWriter, req *http.Request) {
	pkey := req.URL.Query().Get("pkeyFrom")
	accTo := req.URL.Query().Get("accTo")
	amount := req.URL.Query().Get("amount")
	if accTo == "" || pkey == "" || amount == "" {
		fmt.Fprintf(w, "%s\n", "Error: missing 'pkeyFrom' or 'accTo' or 'amount' parameter")
		return
	}
	s, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		fmt.Fprintf(w, "%s: %s\n", "Error: failed to parse 'amount' parameter as float64", err.Error())
		return
	}
	if s <= 0 {
		fmt.Fprintf(w, "%s\n", "Error:  'amount' should be greater than zero")
		return
	}
	resp, err := cl.SendFunds(pkey, accTo, s)
	if err != nil {
		resp = "Error: " + err.Error()
	}
	fmt.Fprintf(w, "%s\n", resp)
}

func chainID(w http.ResponseWriter, req *http.Request) {
	resp, err := cl.ChainID()
	if err != nil {
		resp = "Error: " + err.Error()
	}
	fmt.Fprintf(w, "%s\n", resp)
}

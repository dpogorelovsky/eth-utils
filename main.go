package main

import (
	"fmt"
	"info/client"
	"net/http"
	"strconv"
)

var cl *client.Client

func main() {

	cl = client.New("http://localhost:7545")

	http.HandleFunc("/latestBlock", latestBlock)
	http.HandleFunc("/last10tx", last10tx)
	http.HandleFunc("/balance", balance)
	http.HandleFunc("/sendEth", sendEth)

	fmt.Println("listening to the port :12345")
	http.ListenAndServe(":12345", nil)
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

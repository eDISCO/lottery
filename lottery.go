package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	btc "github.com/vsergeev/btckeygenie/btckey"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func byteString(b []byte) (s string) {
	s = ""
	for i := 0; i < len(b); i++ {
		s += fmt.Sprintf("%02X", b[i])
	}
	return s
}
func main() {
	no_of_keys := 100
	ch := make(chan string)
	quit := make(chan int)
	f, err := os.Open("jackpot.log")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for i := 0; i < no_of_keys; i++ {
		go loop(ch, quit)
	}
	go func() {
		// race condition here, but who would bother..
		select {
		case win := <-ch:
			f.WriteString(win)

		}
	}()
	// wait for goroutines to finish
	for i := 0; i < no_of_keys; i++ {
		<-quit
	}
	close(ch)
}
func loop(ch chan string, quit chan int) {

	priv, err := btc.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("Generating keypair: %s\n", err)
	}
	pri_bytes := priv.ToBytes()
	pri_bytes_b64 := base64.StdEncoding.EncodeToString(pri_bytes)
	/* Convert to Compressed Address */
	address_compressed := priv.ToAddress()
	/* Convert to Uncompressed Address */
	address_uncompressed := priv.ToAddressUncompressed()

	// check the balance of generated addresses:
	var com_balance, uncom_balance int
	com_balance, err = check_balance(address_compressed)
	if err != nil {
		log.Fatalf("Checking balance (comp): %s\n", err)
	}
	uncom_balance, err = check_balance(address_uncompressed)
	if err != nil {
		log.Fatalf("Checking balance (uncomp): %s\n", err)
	}
	if uncom_balance > 0 || com_balance > 0 {
		var balance int
		if uncom_balance > com_balance {
			balance = uncom_balance
		} else {
			balance = com_balance
		}
		ret_str := "Private Key (Base64); " + pri_bytes_b64 + "; Balance;" + strconv.Itoa(balance) + ";"
		ch <- ret_str
	}
	quit <- 1
}
func check_balance(address string) (int, error) {
	query_comp := "https://blockchain.info/q/addressbalance/" + address
	resp, err := http.Get(query_comp)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	bodystring := string(body)
	return strconv.Atoi(bodystring)
}

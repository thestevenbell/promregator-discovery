package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// example run command:  `run main.go -targetUrl=http://localhost:8080/discovery -interval=2`

func main(){
	var TargetUrl = flag.String("targetUrl", "http://localhost:8080", "Discovery URL of the Promregator target to be scraped.")
	var IntervalSeconds = flag.Int( "interval", 30, "Provide the scrape interval in seconds.")

	flag.Parse()
	fmt.Println("Process started with targetUrl: ", *TargetUrl, "and an interval of ", *IntervalSeconds, ".")

	ticker := time.NewTicker(time.Duration(*IntervalSeconds) * time.Second)
	//defer ticker.Stop()

	go func() {
		for t := range ticker.C {
			fmt.Println("Tick at", t)

			if *TargetUrl == ""{
			// TODO - send error log here and exit
			}else {
				callPromregatorDiscoveryEndpoint(TargetUrl)
			}

		}
	}()


	time.Sleep(10000 * time.Millisecond)
	ticker.Stop()
	fmt.Println("Ticker stopped")

	// add ability to gracefully stop the app
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		ticker.Stop()
		fmt.Println("Ticker stopped")
		os.Exit(0)
	}()

}

func callPromregatorDiscoveryEndpoint(targetUrl *string) {
	fmt.Println(*targetUrl)

	resp, err := http.Get(*targetUrl)
	if err != nil {
		fmt.Printf("Error while calling: %s Error message: %s", *targetUrl, err)
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error message:", err)
		} else {
			fmt.Println("body:", string(body[0:20]))
			if !validateResponse(body){
				println("The response from the Promregator discovery endpoint was malformed or missing.  The" +
					"target configuration will not be updated.")
			} else {
				saveResponseToFile(body)
			}
		}
	}
}

func validateResponse(body []byte) bool {
	type targetsJson struct {
		Targets []string
		Labels map[string]string
	}

	var bodyJson []targetsJson

	if err := json.Unmarshal(body, &bodyJson); err != nil {
		fmt.Println(err)
		//panic(err)
		return false
	} else {
		return true
	}
}


func saveResponseToFile(body []byte){
	err := ioutil.WriteFile("/root/data/prometheus-prometheus.json", body, 0644)
	if err != nil{
		fmt.Println("An error occured while attempting to save the targets to file", err)
	}
}




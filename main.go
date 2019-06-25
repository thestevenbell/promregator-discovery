package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// build with `go build .` from within the the project main directory, this builds the binary and save it in in current dir.
// example run command:  `go run main.go -targetUrl=http://localhost:8080/discovery -interval=2 -fileDestination=./data.json`

func main() {

	// create a WaitGroup that will only be Done when a SIG is detected. so that the process does not exit.
	var wg sync.WaitGroup
	wg.Add(1)

	// add ability to gracefully stop the app
	var gracefulStop= make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		defer wg.Done()
		os.Exit(0)
	}()

	var TargetUrl= flag.String("targetUrl", "", "Discovery URL of the Promregator target to be scraped.")
	var IntervalSeconds= flag.Int("interval", 1, "Provide the scrape interval in seconds.")
	var FileDestination= flag.String("fileDestination", "", "Path and filename the Prometheus target output file.")

	flag.Parse()

	if *TargetUrl == "" {
		println("Exiting, no targetUrl was given as a command line argument.")
		os.Exit(0)
	}

	if *IntervalSeconds <= 0 {
		println("Exiting, no interval was given as a command line argument or the value was a non-positive integer.")
		os.Exit(0)
	}

	if *FileDestination == "" {
		println("Exiting, no fileDestination was given as a command line argument.")
		os.Exit(0)
	}

	fmt.Println("Process started with targetUrl: ", *TargetUrl, ", an interval of ", *IntervalSeconds, "and a file path and name of ", *FileDestination, ".")

	ticker := time.NewTicker(time.Duration(*IntervalSeconds) * time.Second)
	defer ticker.Stop()

	go func() {
		for t := range ticker.C {
			fmt.Println("Tick at", t)
			response, err := callPromregatorDiscoveryEndpoint(TargetUrl)
			if err != nil {
				fmt.Println("callPromregatorDiscoveryEndpoint() failed.")
			} else {
				err := validateResponse(response)
				if err != nil {
					fmt.Println("Validation failed.  Not saving configuration.")
				} else {
					saveResponseToFile(response, *FileDestination)
				}
			}
		}
	}()

	// wait forever while the ticker ticks
	wg.Wait()
}

func callPromregatorDiscoveryEndpoint(targetUrl *string) (body []byte, err error) {
	resp, err := http.Get(*targetUrl)
	if err != nil {
		fmt.Printf("Error while calling: %s Error message: %s", *targetUrl, err)
	} else {
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("callPromregatorDiscoveryEndpoint: Error message:", err)
			return body, err
		} else {
			fmt.Println("callPromregatorDiscoveryEndpoint: body:", string(body[0:20]))
		}
	}
	return body, err
}

func validateResponse(body []byte)(err error){
	type targetsJson struct {
		Targets []string
		Labels  map[string]string
	}

	var bodyJson []targetsJson

	if err := json.Unmarshal(body, &bodyJson); err != nil {
		fmt.Println("validateResponse:json.Unmarshal: Error message:", err)
		return err
	}

	if (len(bodyJson) == 0 || len(bodyJson) == 1 ) {
		err := errors.New("The discovery API returned 0 or 1 target scrape endpoints.  This often indicates a problem" +
			" with Promregator's ability to connect with the Cloud Foundry API.  To avoid removing the current Prometheus" +
			" configurations, the configuration file will not be overwritten.  This process will continue to attempt to" +
			" fetch the configurations from Promregator at the configured interval.")
		fmt.Print(err)
		return err
	}

	return nil
}

func saveResponseToFile(body []byte, fileDestination string) {
	err := ioutil.WriteFile(fileDestination, body, 0644)
	if err != nil {
		fmt.Println("An error occured while attempting to save the targets to file", err)
	}
}

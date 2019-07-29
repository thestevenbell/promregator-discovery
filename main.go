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
	"strings"
	"sync"
	"syscall"
	"time"
)

// build with `go build .` from within the the project main directory, this builds the binary and save it in in current dir.
// example run command:  `go run dns.go -TargetURL=http://localhost:8080/discovery -interval=2 -fileDestination=./data.json`

type labelArray []string

func (i *labelArray) String() string {
	return "my string representation"
}

func (i *labelArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type discoveryJSON struct {
	Targets    []string
	LabelsJSON map[string]string
}

var targetsJSON []discoveryJSON

func main() {

	// create a WaitGroup that will only be Done when a SIG is detected. so that the process does not exit.
	var wg sync.WaitGroup
	wg.Add(1)

	// add ability to gracefully stop the app
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		defer wg.Done()
		os.Exit(0)
	}()

	var TargetURL = flag.String("targetURL", "", "Discovery URL of the Promregator target to be scraped.")
	var IntervalSeconds = flag.Int("interval", 1, "Provide the scrape interval in seconds.")
	var FileDestination = flag.String("fileDestination", "", "Path and filename the Prometheus target output file.")
	var Labels labelArray
	flag.Var(&Labels, "label", "meta-labels to be added to each target in the output file. Use key:value format.  The key will be added as the label name with \"__meta_promregator_target_\" prepended and the value will be added as the label value as provided. For example, to add a label to specify the Cloud Foundry availability zone of primary provide the flag with the following value '-label availabilityZone primary'. This will add a label '__meta_promregator_target_availabilityZone' with the value 'primary'.")

	flag.Parse()

	if *TargetURL == "" {
		println("Exiting, no TargetURL was given as a command line argument.")
		os.Exit(1)
	}

	if *IntervalSeconds <= 0 {
		println("Exiting, no interval was given as a command line argument or the value was a non-positive integer.")
		os.Exit(1)
	}

	if *FileDestination == "" {
		println("Exiting, no fileDestination was given as a command line argument.")
		os.Exit(1)
	}

	fmt.Printf("len=%d cap=%d %v\n", len(Labels), cap(Labels), Labels)

	var mapOfLabels = make(map[string]string)

	for _, v := range Labels {
		s := strings.Split(v, ":")
		if len(s) != 2 {
			fmt.Printf("The value provided for the -label flag < %s > should be given as a key value pair delimited with a : semicolon. Exiting", v)
			os.Exit(0)
		}
		fmt.Println("s", s)
		mapOfLabels[s[0]] = s[1]
	}

	fmt.Println("map:", mapOfLabels)

	j, err := json.Marshal(mapOfLabels)

	fmt.Println("json:", string(j), err)

	fmt.Println("Process started with TargetURL: ", *TargetURL, ", an interval of ", *IntervalSeconds, ", a file path of ", *FileDestination, ". The following labels will be applied: ", string(j), ".")

	ticker := time.NewTicker(time.Duration(*IntervalSeconds) * time.Second)
	defer ticker.Stop()

	go func() {
		for t := range ticker.C {
			fmt.Println("Tick at", t)
			response, err := callPromregatorDiscoveryEndpoint(TargetURL)
			if err != nil {
				fmt.Println("callPromregatorDiscoveryEndpoint() failed.")
			} else {
				responseAsJSON, err := validateResponse(response)
				if err != nil {
					fmt.Println("Validation failed.  Not saving configuration.")
				} else {
					addLabels(responseAsJSON, mapOfLabels)
					saveResponseToFile(responseAsJSON, *FileDestination)
				}
			}
		}
	}()

	// wait forever while the ticker ticks
	wg.Wait()
}

func callPromregatorDiscoveryEndpoint(TargetURL *string) (body []byte, err error) {
	resp, err := http.Get(*TargetURL)
	if err != nil {
		fmt.Printf("Error while calling: %s Error message: %s", *TargetURL, err)
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

func validateResponse(body []byte) (responseAsJSON []discoveryJSON, err error) {

	if err := json.Unmarshal(body, &responseAsJSON); err != nil {
		fmt.Println("validateResponse:json.Unmarshal: Error message:", err)
		return responseAsJSON, err
	}

	if len(responseAsJSON) == 0 || len(responseAsJSON) == 1 {
		err := errors.New("The discovery API returned 0 or 1 target scrape endpoints.  This often indicates a problem" +
			" with Promregator's ability to connect with the Cloud Foundry API.  To avoid removing the current Prometheus" +
			" configurations, the configuration file will not be overwritten.  This process will continue to attempt to" +
			" fetch the configurations from Promregator at the configured interval.")
		fmt.Print(err)
		return responseAsJSON, err
	}

	return responseAsJSON, nil
}

func addLabels(responseAsJSON []discoveryJSON, mapOfLabelsToAdd map[string]string) {

	// TODO - print the labels existing
	for i := range responseAsJSON {
		discoveryJSONi := responseAsJSON[i]

		fmt.Printf("responseAsJson.LabelsJSON: %v  .  \n", discoveryJSONi)

		fmt.Println("len(discoveryJSONi.LabelsJSON):", len(discoveryJSONi.LabelsJSON))

		for j, v := range responseAsJSON[i].LabelsJSON {
			fmt.Println("for j := range responseAsJSON[i].LabelsJSON { : index:  ", j, "val:  ", v)
		}
	}

	// TODO - add the labels

}

func saveResponseToFile(responseAsJSON []discoveryJSON, fileDestination string) {

	bytes, err := json.Marshal(responseAsJSON)
	if err != nil {
		panic(err)
	}

	err0 := ioutil.WriteFile(fileDestination, bytes, 0644)
	if err0 != nil {
		fmt.Println("An error occurred while attempting to save the targets to file", err)
	}
}

package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/clientv2/fns"
	"github.com/fnproject/fn_go/provider"
)

// DataMetrics are the data metrics collected throughout the tests
type DataMetrics struct {
	goRoutineMetric string
	elapsedTime     string
	success         bool
	status          string
}

// TestContextCLIFlags represents flags that can be set to identify the app, function and desired concurrency.
// Not changing once the test has started
type TestContextCLIFlags struct {
	AppName            string
	FnName             string
	DesiredConcurrency int
	SleepSeconds       int
	TestDuration       int
}

// FnProviderSetUp used to set up the fn config to be used when invoking concurrent executions
type FnProviderSetUp struct {
	Provider provider.Provider
	Ctx      context.Context
}

// TestContext holds the context of the the test run.
var TestContext TestContextCLIFlags

// RegisterFlags exported
func RegisterFlags() {
	flag.StringVar(&TestContext.AppName, "app", "gosleepapp", "specify app name")
	flag.StringVar(&TestContext.FnName, "fn", "gosleepfn", "specify fn name")
	flag.IntVar(&TestContext.SleepSeconds, "s", 10, "seconds for function to sleep")
	flag.IntVar(&TestContext.DesiredConcurrency, "n", 1, "specify the number of desired concurrent runs of the function specified")
	flag.IntVar(&TestContext.TestDuration, "test-duration", 0, "how long to run test for")
	flag.Parse()
}

// SetUpProvider sets up the fn provider config
func SetUpProvider() *FnProviderSetUp {
	config := provider.NewConfigSourceFromMap(map[string]string{
		"api-url":               "https://localhost:8080",
	})
	currentProvider, err := fn_go.DefaultProviders.ProviderFromConfig("default", config, &provider.NopPassPhraseSource{})
	if err != nil {
		log.Printf("Unable to set up new provider config: %v\n", err)
		os.Exit(1)
	}
	ctx := context.Background()
	return &FnProviderSetUp{
		Provider: currentProvider,
		Ctx:      ctx,
	}
}

// SetUpFn gets the app and func, returning a struct defining the ivoke url to be called
func (f *FnProviderSetUp) SetUpFn(appName string, funcName string) string {
	appClient := f.Provider.APIClientv2().Apps

	appList, err := appClient.ListApps(&apps.ListAppsParams{
		Name:    &appName,
		Context: f.Ctx,
	})
	log.Printf("app: %q\n", appList.Payload.Items[0].Name)
	if err != nil {
		log.Printf("Unable to get app list: %v\n", err)
		os.Exit(1)
	}
	var appID string
	for _, app := range appList.Payload.Items {
		if app.Name == appName {
			appID = app.ID
		} else {
			log.Printf("App not found: %v\n", err)
			os.Exit(1)
		}
	}

	fnClient := f.Provider.APIClientv2().Fns
	fnList, err := fnClient.ListFns(&fns.ListFnsParams{
		Name:    &funcName,
		AppID:   &appID,
		Context: f.Ctx,
	})
	if err != nil {
		log.Printf("Unable to fn list: %v", err)
		os.Exit(1)
	}
	if len(fnList.Payload.Items) != 1 {
		log.Printf("Function not found: %v\n", err)
		os.Exit(1)
	}

	fucntion := fnList.Payload.Items[0]
	log.Printf("function: %q\n", fucntion.Name)
	invokeURL := fucntion.Annotations["fnproject.io/fn/invokeEndpoint"].(string)

	fmt.Println("Setup complete")
	return invokeURL
}

func (f *FnProviderSetUp) deploySleepFn() {

}

func (f *FnProviderSetUp) callSleepFn(i int, invokeURL string, dataChan chan DataMetrics) {
	log.Printf("Go routine %d started.\n", i+1)
	transport := f.Provider.WrapCallTransport(http.DefaultTransport)
	httpClient := http.Client{Transport: transport}

	tNewRequest := time.Now()
	log.Printf("New Request. Time now: %d:%d:%d\n", tNewRequest.Hour(), tNewRequest.Minute(), tNewRequest.Second())
	// Create new request
	req, err := http.NewRequest("POST", invokeURL, strings.NewReader(fmt.Sprintf(`{"seconds": %d}`, TestContext.SleepSeconds)))
	if err != nil {
		log.Printf("Unable to create new post request request: %v", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req = req.WithContext(f.Ctx)

	response, err := httpClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}

	tEnd := time.Now()
	log.Printf("Post request completed. Time now: %d:%d:%d\n", tEnd.Hour(), tEnd.Minute(), tEnd.Second())

	tElapsed := tEnd.Sub(tNewRequest).Round(time.Millisecond) / time.Millisecond
	var statusOrErr string
	success := true
	if err != nil {
		statusOrErr = err.Error()
		log.Printf("Unable to post request: %v", err)
		success = false
	} else if response.StatusCode != 200 {
		log.Printf("Http reponse %d: %v\n", response.StatusCode, err)
		success = false
	} else {
		statusOrErr = response.Status
	}

	log.Printf("Go routine %d ended.\n", i+1)
	dataChan <- DataMetrics{goRoutineMetric: strconv.Itoa(i + 1), elapsedTime: tElapsed.String(), success: success, status: statusOrErr}
}

// RunTest runs concurrent runs (TestContext.DesiredConcurrency) for a duration of time(TestContext.timeDuration)
func RunTest(fileName string) {
	wg := sync.WaitGroup{}
	dataChan := make(chan DataMetrics, TestContext.DesiredConcurrency)

	f := SetUpProvider()
	invokeURL := f.SetUpFn(TestContext.AppName, TestContext.FnName)

	tStart := time.Now()
	for i := 0; i < TestContext.DesiredConcurrency; i++ {
		wg.Add(1)
		i := i
		go func() {
			time.Sleep(time.Duration(float64(TestContext.SleepSeconds)*rand.Float64()) * time.Second)
			for TestContext.TestDuration == 0 || time.Since(tStart) < (time.Duration(TestContext.TestDuration)*time.Second) {
				f.callSleepFn(i, invokeURL, dataChan)
			}
			wg.Done()
		}()

	}

	go func() {
		wg.Wait()
		close(dataChan)
	}()

	file, err := os.OpenFile("resultsConcurrent.csv", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Unable to open file. Creating new file")
		file, err = os.Create(fileName)
		if err != nil {
			log.Printf("Unable to create file: %v\n", err)
			os.Exit(1)
		}
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"test duration(sec.)", "concurrent runs", "fn sleep time(sec.)", "goroutine", "success", "status", "request duration(milli sec.)"})
	for d := range dataChan {
		writer.Write([]string{
			strconv.Itoa(TestContext.TestDuration),
			strconv.Itoa(TestContext.DesiredConcurrency),
			strconv.Itoa(TestContext.SleepSeconds),
			d.goRoutineMetric, strconv.FormatBool(d.success),
			d.status, d.elapsedTime})
		writer.Flush()
	}

	fmt.Println("All Go routines finished executing.")
}

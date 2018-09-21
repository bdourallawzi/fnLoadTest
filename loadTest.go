package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/clientv2/fns"
	"github.com/fnproject/fn_go/provider"
)

// TestContextCLIFlags represents flags that can be set to identify the app, function and desired concurrency
type TestContextCLIFlags struct {
	AppName            string
	FnName             string
	DesiredConcurrency int
}

// FnProviderSetUp used to set up the fn config to be used when invoking concurrent executions
type FnProviderSetUp struct {
	Provider provider.Provider
	Ctx      context.Context
}

// FnAppSetup declares the fn, app and invoke url o be used when invoking concurrent executions
type FnAppSetup struct {
	AppName   string
	FuncName  string
	InvokeURL string
}

// TestContext holds the context of the the test run.
var TestContext TestContextCLIFlags

func registerFlags() {
	flag.StringVar(&TestContext.AppName, "app", "gosleepapp", "specify app name")
	flag.StringVar(&TestContext.FnName, "fn", "gosleepfn", "specify fn name")
	flag.IntVar(&TestContext.DesiredConcurrency, "n", 1, "specify the number of desired concurrent runs of the function specified")
	flag.Parse()
}

// SetUpProvider sets up the fn provider config
func SetUpProvider() *FnProviderSetUp {
	config := provider.NewConfigSourceFromMap(map[string]string{
		"api-url": "http://localhost:8080",
	})
	currentProvider, err := fn_go.DefaultProviders.ProviderFromConfig("default", config, &provider.NopPassPhraseSource{})
	if err != nil {
		fmt.Printf("Unable to set up new provider config: %v\n", err)
		os.Exit(1)
	}
	ctx := context.Background()
	return &FnProviderSetUp{
		Provider: currentProvider,
		Ctx:      ctx,
	}
}

// SetUpFn gets the app and func, returning a struct defining the ivoke url to be called
func (f *FnProviderSetUp) SetUpFn(appName string, funcName string) *FnAppSetup {
	appClient := f.Provider.APIClientv2().Apps

	appList, err := appClient.ListApps(&apps.ListAppsParams{
		Name:    &appName,
		Context: f.Ctx,
	})
	fmt.Printf("app: %q\n", appList.Payload.Items[0].Name)
	if err != nil {
		fmt.Printf("Unable to get app list: %v\n", err)
		os.Exit(1)
	}
	var appID string
	for _, app := range appList.Payload.Items {
		if app.Name == appName {
			appID = app.ID
		} else {
			fmt.Printf("App not found: %v\n", err)
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
		fmt.Printf("Unable to fn list: %v", err)
		os.Exit(1)
	}
	if len(fnList.Payload.Items) != 1 {
		fmt.Printf("Function not found: %v\n", err)
		os.Exit(1)
	}

	fucntion := fnList.Payload.Items[0]
	fmt.Printf("function: %q\n", fucntion.Name)
	invokeURL := fucntion.Annotations["fnproject.io/fn/invokeEndpoint"].(string)

	fmt.Println("Setup complete")
	return &FnAppSetup{
		FuncName:  funcName,
		AppName:   appName,
		InvokeURL: invokeURL,
	}
}

func (f *FnProviderSetUp) callSleepFn(invokeURL string, i int, wg *sync.WaitGroup) {
	fmt.Printf("Go routine %d started.\n", i+1)
	transport := f.Provider.WrapCallTransport(http.DefaultTransport)
	httpClient := http.Client{Transport: transport}
	req, err := http.NewRequest("POST", invokeURL, bytes.NewReader([]byte{}))
	if err != nil {
		fmt.Printf("Unable to create new post request request: %v", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req = req.WithContext(f.Ctx)
	tStart := time.Now()
	fmt.Printf("Function invoked. Time now: %d:%d:%d\n", tStart.Hour(), tStart.Minute(), tStart.Second())
	response, err := httpClient.Do(req)
	tEnd := time.Now()
	fmt.Printf("Post request completed. Time now: %d:%d:%d\n", tEnd.Hour(), tEnd.Minute(), tEnd.Second())
	if err != nil {
		fmt.Printf("Unable to post request: %v", err)
		os.Exit(1)
	}
	if response.StatusCode != 200 {
		fmt.Printf("Http reponse %d: %v\n", response.StatusCode, err)
		os.Exit(1)
	}
	fmt.Printf("Go routine %d ended.\n", i+1)
	wg.Done()
}

func main() {
	registerFlags()

	var wg sync.WaitGroup

	f := SetUpProvider()
	fnAppSetup := f.SetUpFn(TestContext.AppName, TestContext.FnName)
	for i := 0; i < TestContext.DesiredConcurrency; i++ {
		wg.Add(1)
		go f.callSleepFn(fnAppSetup.InvokeURL, i, &wg)
	}

	wg.Wait()
	fmt.Println("All Go routines finished executing.")
}

package main

import (
	"fmt"
	"time"
)

func script3() {
	concurrencyList := []int{10, 20, 30, 40, 50, 60}
	TestContext.TestDuration = 300
	TestContext.SleepSeconds = 10
	for _, concurrency := range concurrencyList {
		TestContext.DesiredConcurrency = concurrency
		RunTest(fmt.Sprintf("results-%dconcurrency-script3.csv", concurrency))
		time.Sleep(time.Second * 30)
	}
}

func script1() {
	concurrencyList := []int{10, 20, 50, 100, 200, 500, 750}
	TestContext.TestDuration = 300
	TestContext.SleepSeconds = 10
	for _, concurrency := range concurrencyList {
		TestContext.DesiredConcurrency = concurrency
		RunTest(fmt.Sprintf("results-%dconcurrency-script1.csv", concurrency))
		time.Sleep(time.Second * 30)
	}
}

func main() {
	RegisterFlags()

	TestContext.DesiredConcurrency = 50
	TestContext.TestDuration = 300
	TestContext.SleepSeconds = 10
	RunTest("results.csv")
}

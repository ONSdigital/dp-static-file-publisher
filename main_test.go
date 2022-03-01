package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ONSdigital/dp-static-file-publisher/features/steps"
	"github.com/cucumber/godog/colors"

	"github.com/ONSdigital/log.go/v2/log"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
)

var (
	componentFlag = flag.Bool("component", false, "perform component tests")
	loggingFlag   = flag.Bool("logging", false, "print logging")
)

type ComponentTest struct {
	MongoFeature *componenttest.MongoFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	if !*loggingFlag {
		buf := bytes.NewBufferString("")
		log.SetDestination(buf, buf)
	}

	component := steps.NewFilePublisherComponent()

	apiFeature := componenttest.NewAPIFeature(component.Initialiser)
	component.ApiFeature = apiFeature

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		component.Reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, e error) (context.Context, error) {
		err := component.Close()
		if err != nil {
			log.Error(ctx, "error closing service", err)
		}
		return ctx, nil
	})

	apiFeature.RegisterSteps(ctx)
	component.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {

}

func TestComponent(t *testing.T) {
	if *componentFlag {
		f := &ComponentTest{}

		status := godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options: &godog.Options{
				Output: colors.Colored(os.Stdout),
				Format: "pretty",
				Paths:  flag.Args(),
			}}.Run()

		fmt.Println("=================================")
		fmt.Printf("Component test coverage: %.2f%%\n", testing.Coverage()*100)
		fmt.Println("=================================")

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}

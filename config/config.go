// Package config contains the implementations of initializing the app properties used by the app
package config

import (
	"context"
	"fmt"
	"os"
	"sync"

	props "github.com/magiconair/properties"
)

var (
	// onceInit guarantees initialization of properties only once
	onceInit      = new(sync.Once)
	concreteImpls = make(map[string]any)
)

const (
	envDetectorKey = "appEnv"
	propsImplKey   = "propsImpl"
)

// Load is an exported method that loads props depending on environment
func Load(ctx context.Context, dir string) error {
	var appErr error
	onceInit.Do(func() {
		appErr = loadImpls(ctx, dir)
	})
	return appErr
}

// GetAll gives all the application properties loaded depending on the env.
// This can be used from anywhere in the app to get configuration data.
func GetAll() *props.Properties {
	x, _ := concreteImpls[propsImplKey].(*props.Properties)
	return x
}

func loadImpls(ctx context.Context, dir string) error {
	if concreteImpls[propsImplKey] == nil {
		allProps, err := load(ctx, dir)
		if err != nil {
			return fmt.Errorf("failed to load properties - %w", err)
		}
		concreteImpls[propsImplKey] = allProps
	}
	return nil
}

func load(_ context.Context, resourceDir string) (p *props.Properties, err error) {
	filesToLoad := []string{resourceDir + "/app.properties"}
	if v, found := os.LookupEnv(envDetectorKey); found {
		filesToLoad = append(filesToLoad, resourceDir+"/app-"+v+".properties")
	}
	p, err = props.LoadFiles(filesToLoad, props.UTF8, false)
	return
}

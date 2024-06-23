package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Plugin interface {
	Name() string
	Initialize() error
	Execute() error
}

var loadedPlugins []Plugin

func loadPlugins() {
	pluginDir := "./plugins"

	err := filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(info.Name()) == ".go" {
			// Assuming that plugins are Go files that implement Plugin interface
			plug, err := loadPlugin(path)
			if err != nil {
				fmt.Println("Error loading plugin:", err)
				return nil
			}

			loadedPlugins = append(loadedPlugins, plug)
			fmt.Println("Loaded plugin:", plug.Name())
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error walking plugin directory:", err)
	}
}

func loadPlugin(path string) (Plugin, error) {
	// You need to implement the logic to load the plugin from Go file
	// This is just a placeholder function
	return nil, fmt.Errorf("loading Go plugins dynamically is not directly supported in this example")
}

func initializePlugins() {
	for _, plug := range loadedPlugins {
		err := plug.Initialize()
		if err != nil {
			fmt.Println("Error initializing plugin:", plug.Name(), err)
		}
	}
}

func executePlugins() {
	for _, plug := range loadedPlugins {
		err := plug.Execute()
		if err != nil {
			fmt.Println("Error executing plugin:", plug.Name(), err)
		}
	}
}

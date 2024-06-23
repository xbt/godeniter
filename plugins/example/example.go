package example

import "fmt"

type ExamplePlugin struct{}

func (p *ExamplePlugin) Name() string {
	return "ExamplePlugin"
}

func (p *ExamplePlugin) Initialize() error {
	fmt.Println("Initializing Example Plugin")
	return nil
}

func (p *ExamplePlugin) Execute() error {
	fmt.Println("Executing Example Plugin")
	return nil
}

func New() *ExamplePlugin {
	return &ExamplePlugin{}
}

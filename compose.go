package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

func main() {
	ctx := context.Background()

	// Load the compose.yml from current directory
	options, err := cli.NewProjectOptions(
		[]string{"compose.yml"},
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithName("myproject"),
	)
	if err != nil {
		log.Fatalf("failed to create project options: %v", err)
	}

	project, err := cli.ProjectFromOptions(ctx, options)
	if err != nil {
		log.Fatalf("failed to load project: %v", err)
	}

	printProject(project)
}

func printProject(p *types.Project) {
	fmt.Printf("Project: %s\n\n", p.Name)

	fmt.Println("Services:")
	for name, svc := range p.Services {
		fmt.Printf("  - %s (image: %s)\n", name, svc.Image)

		for _, port := range svc.Ports {
			fmt.Printf("      port: %s:%s -> %d\n", port.HostIP, port.Published, port.Target)
		}

		if len(svc.DependsOn) > 0 {
			fmt.Printf("      depends_on: ")
			for dep := range svc.DependsOn {
				fmt.Printf("%s ", dep)
			}
			fmt.Println()
		}

		if len(svc.Profiles) > 0 {
			fmt.Printf("      profiles: %v\n", svc.Profiles)
		}
	}

	fmt.Printf("\nNetworks: %v\n", networkNames(p))
	fmt.Printf("Volumes:  %v\n", volumeNames(p))

	// Check if web service is accessible
	if web, ok := p.Services["web"]; ok {
		for _, port := range web.Ports {
			fmt.Printf("\nWeb service reachable at: http://localhost:%s\n", port.Published)
		}
	}

	_ = os.Stdout.Sync()
}

func networkNames(p *types.Project) []string {
	names := make([]string, 0, len(p.Networks))
	for n := range p.Networks {
		names = append(names, n)
	}
	return names
}

func volumeNames(p *types.Project) []string {
	names := make([]string, 0, len(p.Volumes))
	for v := range p.Volumes {
		names = append(names, v)
	}
	return names
}

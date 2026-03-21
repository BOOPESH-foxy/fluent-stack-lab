package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
)

func main() {
	ctx := context.Background()

	// 1. Build Docker CLI client first (needed for validation + service)
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		log.Fatalf("new docker cli: %v", err)
	}
	if err := dockerCli.Initialize(flags.NewClientOptions()); err != nil {
		log.Fatalf("init docker cli: %v", err)
	}

	svc := compose.NewComposeService(dockerCli)

	// 2. Load and parse compose.yml (schema + type validation happens here)
	project, err := loadProject(ctx)
	if err != nil {
		log.Fatalf("compose.yml is invalid: %v", err)
	}

	// 3. Full SDK validation via Config — renders the project and validates
	//    all service references, networks, volumes, depends_on, etc.
	if err := validateProject(ctx, svc, project); err != nil {
		log.Fatalf("compose.yml validation failed:\n%v", err)
	}

	fmt.Println("compose.yml is valid\n")
	printProject(project)

	// 4. Bring the stack up
	fmt.Println("Bringing stack up...")
	_ = svc.Down(ctx, project.Name, api.DownOptions{RemoveOrphans: true, Volumes: false})

	// Only start services that have no profiles (skip adminer which needs --profile tools)
	activeProject, err := project.WithSelectedServices([]string{"db", "web"})
	if err != nil {
		log.Fatalf("filter services: %v", err)
	}

	if err = svc.Create(ctx, activeProject, api.CreateOptions{RemoveOrphans: true}); err != nil {
		log.Fatalf("compose create: %v", err)
	}
	if err = svc.Start(ctx, activeProject.Name, api.StartOptions{Project: activeProject}); err != nil {
		log.Fatalf("compose start: %v", err)
	}

	// 5. Show running containers
	fmt.Println("\nRunning services:")
	containers, err := svc.Ps(ctx, activeProject.Name, api.PsOptions{All: false})
	if err != nil {
		log.Fatalf("compose ps: %v", err)
	}
	for _, c := range containers {
		fmt.Printf("  %-30s  status: %s\n", c.Name, c.State)
	}

	fmt.Println("\nWeb available at: http://localhost:8070")
}

// validateProject uses the SDK's Config method to do a full validation pass:
// - resolves all service image/build references
// - validates network and volume references
// - checks depends_on targets exist
// - validates port bindings and environment variables
func validateProject(ctx context.Context, svc api.Compose, project *types.Project) error {
	_, err := svc.Config(ctx, project, api.ConfigOptions{
		Format:              "yaml",
		ResolveImageDigests: false,
	})
	return err
}

func loadProject(ctx context.Context) (*types.Project, error) {
	options, err := cli.NewProjectOptions(
		[]string{"compose.yml"},
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithName("myproject"),
		cli.WithConsistency(true), // enforce spec consistency checks
	)
	if err != nil {
		return nil, err
	}
	return cli.ProjectFromOptions(ctx, options)
}

func printProject(p *types.Project) {
	fmt.Printf("Project:  %s\n", p.Name)
	fmt.Printf("Services: %d\n", len(p.Services))
	for name, s := range p.Services {
		image := s.Image
		if image == "" {
			image = "(build)"
		}
		profiles := ""
		if len(s.Profiles) > 0 {
			profiles = fmt.Sprintf(" [profiles: %v]", s.Profiles)
		}
		deps := ""
		if len(s.DependsOn) > 0 {
			for d := range s.DependsOn {
				deps += d + " "
			}
			deps = " depends_on: " + deps
		}
		fmt.Printf("  %-12s image: %-25s%s%s\n", name, image, deps, profiles)
	}
	fmt.Printf("Networks: %v\n", networkNames(p))
	fmt.Printf("Volumes:  %v\n", volumeNames(p))
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

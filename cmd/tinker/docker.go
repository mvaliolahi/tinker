package main

import (
        "fmt"
        "strings"

        "github.com/mvaliolahi/tinker/internal/detect"
        "github.com/mvaliolahi/tinker/internal/ui"
        "github.com/spf13/cobra"
)

func dockerCmd() *cobra.Command {
        cmd := &cobra.Command{
                Use:   "docker",
                Short: "Inspect Docker Compose services",
        }

        cmd.AddCommand(dockerListCmd(), dockerInfoCmd())

        cmd.RunE = func(_ *cobra.Command, _ []string) error {
                dir, err := resolveDir()
                if err != nil {
                        return err
                }

                result := detect.New(dir).Detect()
                if result.Docker == nil {
                        fmt.Println(ui.Warning("No Docker Compose file found"))
                        fmt.Println(ui.Hint("Add a docker-compose.yml to your project root"))
                        return nil
                }

                printDockerInfo(result.Docker)
                return nil
        }

        return cmd
}

func dockerListCmd() *cobra.Command {
        return &cobra.Command{
                Use:     "list",
                Aliases: []string{"ls"},
                Short:   "List Docker Compose services",
                RunE: func(_ *cobra.Command, _ []string) error {
                        dir, err := resolveDir()
                        if err != nil {
                                return err
                        }

                        result := detect.New(dir).Detect()
                        if result.Docker == nil {
                                fmt.Println(ui.Warning("No Docker Compose file found"))
                                return nil
                        }

                        fmt.Println()
                        fmt.Println("  " + ui.Bold("Docker Compose Services"))
                        fmt.Println(ui.KeyValue("file", result.Docker.ComposeFile))
                        fmt.Println()

                        for _, svc := range result.Docker.Services {
                                name := ui.Accent(svc.Name)
                                details := ""
                                if svc.Image != "" {
                                        details = ui.Dim(svc.Image)
                                }
                                if svc.Ports != "" {
                                        if details != "" {
                                                details += " "
                                        }
                                        details += ui.Dim("ports: " + svc.Ports)
                                }
                                fmt.Printf("  %-20s %s\n", name, details)

                                // Show detected service types
                                var types []string
                                if svc.HasDB {
                                        types = append(types, ui.DBLabel()+" database")
                                }
                                if svc.HasAPI {
                                        types = append(types, ui.APILabel()+" api")
                                }
                                if svc.HasGRPC {
                                        types = append(types, ui.GRPCLabel()+" grpc")
                                }
                                if len(types) > 0 {
                                        fmt.Printf("    detected: %s\n", strings.Join(types, ", "))
                                }
                        }
                        fmt.Println()
                        return nil
                },
        }
}

func dockerInfoCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "info",
                Short: "Show Docker Compose service details",
                RunE: func(_ *cobra.Command, _ []string) error {
                        dir, err := resolveDir()
                        if err != nil {
                                return err
                        }

                        result := detect.New(dir).Detect()
                        if result.Docker == nil {
                                fmt.Println(ui.Warning("No Docker Compose file found"))
                                return nil
                        }

                        printDockerInfo(result.Docker)
                        return nil
                },
        }
}

func printDockerInfo(d *detect.DockerResult) {
        fmt.Println()
        fmt.Println("  " + ui.Bold("Docker Compose"))
        fmt.Println(ui.KeyValue("file", d.ComposeFile))
        fmt.Println(ui.KeyValue("services", fmt.Sprintf("%d", len(d.Services))))
        fmt.Println()

        for _, svc := range d.Services {
                fmt.Println("  " + ui.Accent(svc.Name))
                if svc.Image != "" {
                        fmt.Println(ui.KeyValue("image", svc.Image))
                }
                if svc.Ports != "" {
                        fmt.Println(ui.KeyValue("ports", svc.Ports))
                }

                var capabilities []string
                if svc.HasDB {
                        capabilities = append(capabilities, "database")
                }
                if svc.HasAPI {
                        capabilities = append(capabilities, "api")
                }
                if svc.HasGRPC {
                        capabilities = append(capabilities, "grpc")
                }
                if len(capabilities) > 0 {
                        fmt.Println(ui.KeyValue("type", fmt.Sprintf("%v", capabilities)))
                }
                fmt.Println()
        }
}

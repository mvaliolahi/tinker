package main

import (
	"fmt"
	"strings"

	"github.com/mvaliolahi/tinker/internal/api"
	"github.com/mvaliolahi/tinker/internal/ui"
)

// runAPIExplorer runs a simple interactive loop for exploring the API spec.
func runAPIExplorer(s *api.Session, spec *api.Spec) error {
	fmt.Println()
	fmt.Println("  " + ui.APILabel() + " " + ui.Bold("API Explorer"))
	fmt.Println(ui.KeyValue("base", s.BaseURL))
	if spec.Title != "" {
		fmt.Println(ui.KeyValue("spec", spec.Title+" v"+spec.Version))
	}
	fmt.Println(ui.KeyValue("endpoints", fmt.Sprintf("%d available", len(spec.Endpoints))))
	fmt.Println()
	fmt.Println(ui.Dim("  Commands:"))
	fmt.Println(ui.Dim("    list              List all endpoints"))
	fmt.Println(ui.Dim("    call <method> <path> [body]   Call an endpoint"))
	fmt.Println(ui.Dim("    find <keyword>    Search endpoints"))
	fmt.Println(ui.Dim("    tags              List tags"))
	fmt.Println(ui.Dim("    quit / q          Exit"))
	fmt.Println()

	for {
		fmt.Print(ui.Accent("tinker/api> "))

		var input string
		fmt.Scanln(&input)
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "quit", "q", "exit":
			fmt.Println(ui.Dim("  Bye!"))
			return nil

		case "list", "ls", "endpoints":
			printEndpointsByTag(spec)

		case "tags":
			byTag := spec.EndpointsByTag()
			fmt.Println()
			for tag, endpoints := range byTag {
				fmt.Println(ui.Bullet(tag, fmt.Sprintf("%d endpoint(s)", len(endpoints))))
			}
			fmt.Println()

		case "find", "search":
			if len(parts) < 2 {
				fmt.Println(ui.Dim("  Usage: find <keyword>"))
				continue
			}
			keyword := strings.ToLower(strings.Join(parts[1:], " "))
			var matches []api.SpecEndpoint
			for _, ep := range spec.Endpoints {
				if strings.Contains(strings.ToLower(ep.Path), keyword) ||
					strings.Contains(strings.ToLower(ep.Summary), keyword) ||
					strings.Contains(strings.ToLower(ep.OperationID), keyword) {
					matches = append(matches, ep)
				}
			}
			if len(matches) == 0 {
				fmt.Println(ui.Dim("  No endpoints matching: " + keyword))
			} else {
				fmt.Println()
				printEndpoints(matches)
			}

		case "call":
			if len(parts) < 3 {
				fmt.Println(ui.Dim("  Usage: call <method> <path> [body]"))
				continue
			}
			method := strings.ToUpper(parts[1])
			path := parts[2]
			body := ""
			if len(parts) > 3 {
				body = strings.Join(parts[3:], " ")
			}

			// Look up endpoint info from spec
			ep := spec.FindEndpoint(method, path)
			fmt.Println()
			fmt.Println("  " + ui.APILabel() + " " + ui.HTTPMethod(method) + " " + ui.Accent(path))
			if ep != nil && ep.Summary != "" {
				fmt.Println(ui.KeyValue("summary", ep.Summary))
			}
			fmt.Println(ui.KeyValue("url", s.BaseURL+path))
			if body != "" {
				fmt.Println(ui.KeyValue("body", body))
			}
			fmt.Println()

			out, err := s.Request(method, path, body, nil)
			fmt.Print(out)
			if err != nil {
				fmt.Println(ui.Error(err.Error()))
			}

		case "help", "?":
			fmt.Println()
			fmt.Println(ui.Dim("  list              List all endpoints"))
			fmt.Println(ui.Dim("    call <method> <path> [body]   Call an endpoint"))
			fmt.Println(ui.Dim("    find <keyword>    Search endpoints"))
			fmt.Println(ui.Dim("    tags              List tags"))
			fmt.Println(ui.Dim("    quit              Exit"))
			fmt.Println()

		default:
			// Try to interpret as a quick API call: GET /users, POST /users '{}'
			method := strings.ToUpper(parts[0])
			switch method {
			case "GET", "POST", "PUT", "PATCH", "DELETE":
				if len(parts) < 2 {
					fmt.Println(ui.Dim("  Usage: " + method + " <path> [body]"))
					continue
				}
				path := parts[1]
				body := ""
				if len(parts) > 2 {
					body = strings.Join(parts[2:], " ")
				}

				ep := spec.FindEndpoint(method, path)
				fmt.Println()
				fmt.Println("  " + ui.APILabel() + " " + ui.HTTPMethod(method) + " " + ui.Accent(path))
				if ep != nil && ep.Summary != "" {
					fmt.Println(ui.KeyValue("summary", ep.Summary))
				}
				fmt.Println(ui.KeyValue("url", s.BaseURL+path))
				if body != "" {
					fmt.Println(ui.KeyValue("body", body))
				}
				fmt.Println()

				out, err := s.Request(method, path, body, nil)
				fmt.Print(out)
				if err != nil {
					fmt.Println(ui.Error(err.Error()))
				}

			default:
				fmt.Println(ui.Dim("  Unknown command. Type 'help' for available commands."))
			}
		}
	}
}

// printEndpoints prints a flat list of endpoints with method badges.
func printEndpoints(endpoints []api.SpecEndpoint) {
	for _, ep := range endpoints {
		method := ui.HTTPMethod(ep.Method)
		summary := ""
		if ep.Summary != "" {
			summary = ui.Dim(" — " + ep.Summary)
		}
		fmt.Printf("  %s  %-30s%s\n", method, ui.Accent(ep.Path), summary)
	}
	fmt.Println()
}

// printEndpointsByTag prints endpoints grouped by their first tag.
func printEndpointsByTag(spec *api.Spec) {
	byTag := spec.EndpointsByTag()

	// Sort tags for deterministic output
	tags := make([]string, 0, len(byTag))
	for t := range byTag {
		tags = append(tags, t)
	}

	for _, tag := range tags {
		endpoints := byTag[tag]
		fmt.Println("  " + ui.Bold(tag))
		printEndpoints(endpoints)
	}
}

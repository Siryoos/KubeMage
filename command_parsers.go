package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseGenDeployCommand(input string) (DeploymentOptions, error) {
	var opts DeploymentOptions
	fields := strings.Fields(input)
	if len(fields) < 2 {
		return opts, fmt.Errorf("usage: /gen-deploy <name> --image <image> [--replicas N] [--port P]")
	}

	opts.Name = fields[1]

	for i := 2; i < len(fields); i++ {
		switch fields[i] {
		case "--image":
			i++
			if i >= len(fields) {
				return opts, fmt.Errorf("missing value for --image")
			}
			opts.Image = fields[i]
		case "--replicas":
			i++
			if i >= len(fields) {
				return opts, fmt.Errorf("missing value for --replicas")
			}
			replicas, err := strconv.Atoi(fields[i])
			if err != nil {
				return opts, fmt.Errorf("invalid replicas value: %v", err)
			}
			opts.Replicas = replicas
		case "--port":
			i++
			if i >= len(fields) {
				return opts, fmt.Errorf("missing value for --port")
			}
			port, err := strconv.Atoi(fields[i])
			if err != nil {
				return opts, fmt.Errorf("invalid port value: %v", err)
			}
			opts.Port = port
		default:
			return opts, fmt.Errorf("unknown flag %s", fields[i])
		}
	}

	if strings.TrimSpace(opts.Image) == "" {
		return opts, fmt.Errorf("--image is required")
	}

	return opts, nil
}

func parseGenHelmCommand(input string) (HelmChartOptions, error) {
	var opts HelmChartOptions
	fields := strings.Fields(input)
	if len(fields) < 2 {
		return opts, fmt.Errorf("usage: /gen-helm <chartName> [--description DESC] [--version VER] [--app-version APP]")
	}

	opts.Name = fields[1]

	for i := 2; i < len(fields); i++ {
		switch fields[i] {
		case "--description":
			i++
			if i >= len(fields) {
				return opts, fmt.Errorf("missing value for --description")
			}
			opts.Description = fields[i]
		case "--version":
			i++
			if i >= len(fields) {
				return opts, fmt.Errorf("missing value for --version")
			}
			opts.Version = fields[i]
		case "--app-version":
			i++
			if i >= len(fields) {
				return opts, fmt.Errorf("missing value for --app-version")
			}
			opts.AppVersion = fields[i]
		default:
			return opts, fmt.Errorf("unknown flag %s", fields[i])
		}
	}

	return opts, nil
}

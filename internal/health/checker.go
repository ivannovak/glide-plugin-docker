package health

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ivannovak/glide/internal/context"
	"github.com/ivannovak/glide-plugin-docker/internal/container"
	"github.com/ivannovak/glide-plugin-docker/internal/resolver"
)

// HealthStatus represents the health status of a container
type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"
	HealthUnhealthy HealthStatus = "unhealthy"
	HealthStarting  HealthStatus = "starting"
	HealthNone      HealthStatus = "none"
	HealthUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a container health check result
type HealthCheck struct {
	Service      string       `json:"service"`
	Container    string       `json:"container"`
	Status       HealthStatus `json:"status"`
	FailingCount int          `json:"failing_count"`
	Log          []string     `json:"log"`
	LastChecked  time.Time    `json:"last_checked"`
	Error        string       `json:"error,omitempty"`
}

// ServiceHealth represents the overall health of a service
type ServiceHealth struct {
	Service    string        `json:"service"`
	Healthy    bool          `json:"healthy"`
	Containers []HealthCheck `json:"containers"`
	Summary    string        `json:"summary"`
}

// HealthMonitor monitors container health
type HealthMonitor struct {
	ctx      *context.ProjectContext
	resolver *resolver.Resolver
	manager  *container.ContainerManager
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(ctx *context.ProjectContext) *HealthMonitor {
	return &HealthMonitor{
		ctx:      ctx,
		resolver: resolver.NewResolver(ctx),
		manager:  container.NewContainerManager(ctx),
	}
}

// CheckHealth checks the health of all containers
func (hm *HealthMonitor) CheckHealth() ([]ServiceHealth, error) {
	// Get all containers
	containers, err := hm.manager.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get container status: %w", err)
	}

	// Group containers by service
	serviceMap := make(map[string][]container.Container)
	for _, c := range containers {
		serviceMap[c.Service] = append(serviceMap[c.Service], c)
	}

	// Check health for each service
	var results []ServiceHealth
	for service, serviceContainers := range serviceMap {
		health := ServiceHealth{
			Service:    service,
			Healthy:    true,
			Containers: []HealthCheck{},
		}

		for _, container := range serviceContainers {
			check := hm.checkContainerHealth(container)
			health.Containers = append(health.Containers, check)

			if check.Status != HealthHealthy && check.Status != HealthNone {
				health.Healthy = false
			}
		}

		// Generate summary
		health.Summary = hm.generateHealthSummary(health)
		results = append(results, health)
	}

	return results, nil
}

// checkContainerHealth checks the health of a single container
func (hm *HealthMonitor) checkContainerHealth(c container.Container) HealthCheck {
	check := HealthCheck{
		Service:     c.Service,
		Container:   c.Name,
		Status:      HealthUnknown,
		LastChecked: time.Now(),
	}

	// Check container state
	if c.State != "running" {
		check.Status = HealthNone
		check.Error = fmt.Sprintf("Container is %s", c.State)
		return check
	}

	// Get health status from Docker
	cmd := exec.Command("docker", "inspect", "--format",
		"{{json .State.Health}}", c.ID)

	output, err := cmd.Output()
	if err != nil {
		check.Error = fmt.Sprintf("Failed to inspect container: %v", err)
		return check
	}

	// Parse health data
	var healthData struct {
		Status       string `json:"Status"`
		FailingCount int    `json:"FailingCount"`
		Log          []struct {
			Output string `json:"Output"`
		} `json:"Log"`
	}

	output = []byte(strings.TrimSpace(string(output)))
	if string(output) == "null" || string(output) == "<no value>" {
		// No health check defined
		check.Status = HealthNone
		return check
	}

	if err := json.Unmarshal(output, &healthData); err != nil {
		check.Error = fmt.Sprintf("Failed to parse health data: %v", err)
		return check
	}

	// Map Docker health status to our status
	switch healthData.Status {
	case "healthy":
		check.Status = HealthHealthy
	case "unhealthy":
		check.Status = HealthUnhealthy
	case "starting":
		check.Status = HealthStarting
	default:
		check.Status = HealthUnknown
	}

	check.FailingCount = healthData.FailingCount

	// Extract log entries
	for _, entry := range healthData.Log {
		if entry.Output != "" {
			check.Log = append(check.Log, entry.Output)
		}
	}

	return check
}

// WaitForHealthy waits for containers to become healthy
func (hm *HealthMonitor) WaitForHealthy(timeout time.Duration, services ...string) error {
	start := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			health, err := hm.CheckHealth()
			if err != nil {
				return fmt.Errorf("failed to check health: %w", err)
			}

			allHealthy := true
			for _, serviceHealth := range health {
				// If specific services requested, only check those
				if len(services) > 0 {
					found := false
					for _, s := range services {
						if s == serviceHealth.Service {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}

				// Check if service is healthy
				if !serviceHealth.Healthy {
					allHealthy = false

					// Check for critical failures
					for _, container := range serviceHealth.Containers {
						if container.Status == HealthUnhealthy && container.FailingCount > 3 {
							return fmt.Errorf("service %s is unhealthy after %d attempts",
								serviceHealth.Service, container.FailingCount)
						}
					}
				}
			}

			if allHealthy {
				return nil
			}

		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for containers to become healthy")
		}

		// Check if we've exceeded timeout
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for containers to become healthy")
		}
	}
}

// generateHealthSummary creates a human-readable summary
func (hm *HealthMonitor) generateHealthSummary(health ServiceHealth) string {
	if len(health.Containers) == 0 {
		return "No containers"
	}

	healthy := 0
	unhealthy := 0
	starting := 0
	stopped := 0

	for _, container := range health.Containers {
		switch container.Status {
		case HealthHealthy, HealthNone:
			healthy++
		case HealthUnhealthy:
			unhealthy++
		case HealthStarting:
			starting++
		default:
			stopped++
		}
	}

	parts := []string{}
	if healthy > 0 {
		parts = append(parts, fmt.Sprintf("%d healthy", healthy))
	}
	if unhealthy > 0 {
		parts = append(parts, fmt.Sprintf("%d unhealthy", unhealthy))
	}
	if starting > 0 {
		parts = append(parts, fmt.Sprintf("%d starting", starting))
	}
	if stopped > 0 {
		parts = append(parts, fmt.Sprintf("%d stopped", stopped))
	}

	if len(parts) == 0 {
		return "Unknown status"
	}

	return strings.Join(parts, ", ")
}

// GetServiceHealth gets health status for a specific service
func (hm *HealthMonitor) GetServiceHealth(service string) (*ServiceHealth, error) {
	allHealth, err := hm.CheckHealth()
	if err != nil {
		return nil, err
	}

	for _, health := range allHealth {
		if health.Service == service {
			return &health, nil
		}
	}

	return nil, fmt.Errorf("service '%s' not found", service)
}

// IsHealthy checks if all services are healthy
func (hm *HealthMonitor) IsHealthy() bool {
	health, err := hm.CheckHealth()
	if err != nil {
		return false
	}

	for _, serviceHealth := range health {
		if !serviceHealth.Healthy {
			return false
		}
	}

	return true
}

// GetUnhealthyServices returns list of unhealthy services
func (hm *HealthMonitor) GetUnhealthyServices() ([]string, error) {
	health, err := hm.CheckHealth()
	if err != nil {
		return nil, err
	}

	var unhealthy []string
	for _, serviceHealth := range health {
		if !serviceHealth.Healthy {
			unhealthy = append(unhealthy, serviceHealth.Service)
		}
	}

	return unhealthy, nil
}

// RestartUnhealthy restarts all unhealthy services
func (hm *HealthMonitor) RestartUnhealthy() error {
	unhealthy, err := hm.GetUnhealthyServices()
	if err != nil {
		return fmt.Errorf("failed to get unhealthy services: %w", err)
	}

	if len(unhealthy) == 0 {
		return nil // Nothing to restart
	}

	return hm.manager.Restart(unhealthy...)
}

// MonitorHealth continuously monitors health and calls callback on changes
func (hm *HealthMonitor) MonitorHealth(interval time.Duration, onChange func([]ServiceHealth)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastHealth []ServiceHealth

	for range ticker.C {
		health, err := hm.CheckHealth()
		if err != nil {
			continue
		}

		// Check if health has changed
		if !hm.healthEqual(lastHealth, health) {
			if onChange != nil {
				onChange(health)
			}
			lastHealth = health
		}
	}
}

// healthEqual compares two health states
func (hm *HealthMonitor) healthEqual(a, b []ServiceHealth) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for comparison
	aMap := make(map[string]bool)
	bMap := make(map[string]bool)

	for _, h := range a {
		aMap[h.Service] = h.Healthy
	}
	for _, h := range b {
		bMap[h.Service] = h.Healthy
	}

	// Compare maps
	for service, healthy := range aMap {
		if bHealthy, exists := bMap[service]; !exists || bHealthy != healthy {
			return false
		}
	}

	return true
}

package cmd

import "time"

// Config struct to hold the list of dependencies
type Config struct {
	Dependencies []Dependency `yaml:"dependencies"`
}

// Dependency struct to hold package version or branch information
type Dependency struct {
	Package string `yaml:"package"`
	Version string `yaml:"version,omitempty"`
	Branch  string `yaml:"branch,omitempty"`
}

type Package struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
}

type Module struct {
	Require []Package `json:"Require"`
}

// Commit is a structure representing the commit response from GitHub API
type Commit struct {
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

// MockCommandExecutor simulates the behavior of the commandExecutor interface
type MockCommandExecutor struct {
	Outcome   string // Outcome to return for Output() method
	RunErr    error  // Error to return for Run() method
	OutputErr error  // Error to return for Output() method
}

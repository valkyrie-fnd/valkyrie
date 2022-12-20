package ops

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

// A Profiles struct holds all loaded profiles of the application.
// A profile can be used at runtime to customize behavior of the application.
// This was inspired by other frameworks, such as Micronaut (https://github.com/micronaut-projects/micronaut-core/blob/v3.6.3/inject/src/main/java/io/micronaut/context/env/DefaultEnvironment.java)
type Profiles struct {
	profiles map[string]bool
}

func NewProfiles() *Profiles {
	return &Profiles{profiles: map[string]bool{}}
}

// List returns the profiles of p
func (p *Profiles) List() []string {
	list := []string{}
	for profile := range p.profiles {
		list = append(list, profile)
	}
	return list
}

// Has returns true if profile is present in p
func (p *Profiles) Has(profile string) bool {
	return p.profiles[profile]
}

// Load profiles explicitly set from VALK_PROFILES=dev,linux,cloud,gcp
//
// Try to deduce profiles unless disabled by VALK_PROFILES_DEDUCE=false
func (p *Profiles) Load() *Profiles {
	if val, found := os.LookupEnv("VALK_PROFILES_DEDUCE"); found {
		deduceProfiles, err := strconv.ParseBool(val)
		if err != nil && deduceProfiles {
			p.deduce()
		}
	} else {
		p.deduce()
	}

	if val, found := os.LookupEnv("VALK_PROFILES"); found {
		for _, profile := range strings.Split(strings.ToLower(val), ",") {
			p.profiles[profile] = true
		}
	}

	return p
}

func (p *Profiles) deduce() {
	if _, found := os.LookupEnv("KUBERNETES_SERVICE_HOST"); found {
		p.profiles["k8s"] = true
	}
	if _, found := os.LookupEnv("GOOGLE_COMPUTE_METADATA"); found {
		p.profiles["gcp"] = true
		p.profiles["cloud"] = true
	}
	if _, err := os.Stat("/.dockerenv"); err == nil {
		p.profiles["docker"] = true
	}
	p.profiles[runtime.GOARCH] = true
	p.profiles[runtime.GOOS] = true
}

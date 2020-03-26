package domain

import "sync"

// NetworkConfig ...
type NetworkConfig struct {
	sync.Mutex
	// TechInterface string
	// FwmarkNumber  string
	// PathToIfcfgTunnelFiles            string
	// SysctlConfFilePath                string
	// PathToKeepalivedConfig            string
	// PathToKeepalivedDConfigConfigured string
	// PathToKeepalivedDConfigEnabled    string
}

// NewNetworkConfig ...
func NewNetworkConfig(
// techInterface,
// fwmarkNumber,
// pathToIfcfgTunnelFiles,
// sysctlConfFilePath,
// pathToKeepalivedConfig,
// pathToKeepalivedDConfigConfigured,
// pathToKeepalivedDConfigEnabled string
) *NetworkConfig {
	return &NetworkConfig{
		// TechInterface: techInterface,
		// FwmarkNumber:  fwmarkNumber,
		// PathToIfcfgTunnelFiles:            pathToIfcfgTunnelFiles,
		// SysctlConfFilePath:                sysctlConfFilePath,
		// PathToKeepalivedConfig:            pathToKeepalivedConfig,
		// PathToKeepalivedDConfigConfigured: pathToKeepalivedDConfigConfigured,
		// PathToKeepalivedDConfigEnabled:    pathToKeepalivedDConfigEnabled,
	}
}

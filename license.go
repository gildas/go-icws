package icws

// License describes a PureConnect License
type License struct {
	Name       string `json:"name"`
	IsAssigned bool   `json:"isAssigned"`
}

type LicenseProperties struct {
	Active             bool                `json:"licenseActive"`
	HasClientAccess    bool                `json:"hasClientAccess"`
	MediaLicense       MediaLicense        `json:"mediaLevel"`
	AllocationType     AllocationLicense   `json:"allocationType"`
	IPALicense         IPALicense          `json:"interactionProcessorAutomationType"`
	AdditionalLicenses []AdditionalLicense `json:"additionalLicenses"`
}

type AdditionalLicense struct {
	ID      string `json:"id"`
	Name    string `json:"displayName"`
	SelfURI string `json:"uri"`
}

type AllocationLicense uint32

const (
	AssignableAllocation AllocationLicense = iota
	ConcurrentAllocation
)

type MediaLicense uint32

const (
	Media1 MediaLicense = iota
	Media2
	Media3Plus
)

type IPALicense uint32

const (
	IPADirectRoutedWorkItems IPALicense = iota
	IPAGroupRoutedWorkItems
	IPAProcessMonitor
	IPAProcessDesigner
)

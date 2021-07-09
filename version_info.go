package icws

import (
	"encoding/json"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

// VersionInfo describes the PureConnect Version
type VersionInfo struct {
	Major          int    `json:"-"`
	Minor          int    `json:"-"`
	Patch          int    `json:"-"`
	Build          int    `json:"-"`
	Product        string `json:"productId"`
	Codebase       string `json:"codebaseId"`
	ProductRelease string `json:"productReleaseDisplayString"`
	ProductPath    string `json:"productPatchDisplayString"`

}

// GetVersion retrieves the PureConnect version
func (session *Session) GetVersion() (*VersionInfo, error) {
	version := VersionInfo{}
	err := session.sendGet("/connection/version", &version)
	return &version, err
}

// UnmarshalJSON unmarshals from JSON
//
// implements json.Unmarshaler
func (version *VersionInfo) UnmarshalJSON(payload []byte) (err error) {
	type surrogate VersionInfo
	var inner struct {
		surrogate
		Major core.FlexInt `json:"majorVersion"`
		Minor core.FlexInt `json:"minorVersion"`
		Patch core.FlexInt `json:"su"`
		Build core.FlexInt `json:"build"`
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*version = VersionInfo(inner.surrogate)
	version.Major = int(inner.Major)
	version.Minor = int(inner.Minor)
	version.Patch = int(inner.Patch)
	version.Build = int(inner.Build)
	return nil
}

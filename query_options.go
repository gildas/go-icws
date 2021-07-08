package icws

import "strings"

type QueryOptions struct {
	Fields                      QueryFieldSelector `json:"select"`
	Where                       QueryConditions    `json:"where"`
	Rights                      QueryRightsFilter  `json:"rightsFilter"`
	ActualValues                bool               `json:"actualValues"`
	InheritedValues             bool               `json:"inheritedValues"`
	SinglePropertyInheritedFrom bool               `json:"singlePropertyInheritedFrom"`
}

type QueryFieldSelector []string
type QueryRightsFilter string
type QueryConditions []string

// AsQueryParameters return a parameter map for session.send
func (options QueryOptions) AsQueryParameters() map[string]string {
	parameters := map[string]string{}
	
	if len(options.Fields) > 0 {
		parameters["select"] = strings.Join(options.Fields, ",")
	}
	if len(options.Where) > 0 {
		parameters["where"] = strings.Join(options.Where, ",")
	}
	if len(options.Rights) > 0 {
		parameters["rightsFilter"] = string(options.Rights)
	}
	if options.ActualValues {
		parameters["actualValues"] = "true"
	}
	if options.InheritedValues {
		parameters["inheritedValues"] = "true"
	}
	if options.SinglePropertyInheritedFrom {
		parameters["singlePropertyInheritedFrom"] = "true"
	}
	return parameters
}

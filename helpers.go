package icws

import (
	"reflect"
)

type Identifiable interface {
	GetID() string
}

// IDList collects the Identifiers of items
//
// 
// If some of the given items are not identifiable, it panics
func IDList(identifiables interface{}) []string {
	// We have to use the reflect package, because Go does not allow casting from
	// []Type to []Identifiable, even if Type implements Identifiable
	identifiableInterface := reflect.TypeOf((*Identifiable)(nil)).Elem()

	switch reflect.TypeOf(identifiables).Kind() {
	case reflect.Slice:
		slice := reflect.ValueOf(identifiables)
		if slice.Len() > 0 && slice.Index(0).CanInterface() && slice.Index(0).Type().Implements(identifiableInterface) {
			items := make([]string, slice.Len())
			for i := 0; i < slice.Len(); i++ {
				items[i] = slice.Index(i).Interface().(Identifiable).GetID()
			}
			return items
		}
	default:
		value := reflect.ValueOf(identifiables)
		if value.CanInterface() && value.Type().Implements(identifiableInterface) {
			return []string{value.Interface().(Identifiable).GetID()}
		}
	}
	return []string{}
}

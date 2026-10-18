package guidefs

import (
	"github.com/Uzer007-admin/granio/v2"
)

type jsonResource struct {
	granio.Resource `json:"-"`
}

func (r *jsonResource) MarshalJSON() ([]byte, error) {
	icon := "\"" + IconName(r.Resource) + "\""

	return []byte(icon), nil
}

// WrapResource wraps a granio.Resource for integration with JSON
func WrapResource(r granio.Resource) granio.Resource {
	return &jsonResource{r}
}

// IconName returns the name for an icon
func IconName(res granio.Resource) string {
	name := res.Name()
	// strip prefix numbers to unwrap
	for name[0] >= '0' && name[0] <= '9' {
		name = name[1:]
	}

	ret, ok := IconReverse[name]
	if !ok {
		return "BrokenImageIcon"
	}

	return ret
}

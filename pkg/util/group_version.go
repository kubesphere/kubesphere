package util

// GroupVersion contains the "group" and the "version", which uniquely identifies the API.
type GroupVersion struct {
	Group   string
	Version string
}

func (gv GroupVersion) String() string {
	if len(gv.Group) == 0 {
		return gv.Version
	}

	return gv.Group + "/" + gv.Version
}

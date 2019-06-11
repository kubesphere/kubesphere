// Get duplication information for a project.
package sonargo

import "net/http"

type DuplicationsService struct {
	client *Client
}

type DuplicationsShowObject struct {
	Duplications []*Duplication `json:"duplications,omitempty"`
	Files        *Files         `json:"files,omitempty"`
}

type Block struct {
	Ref  string `json:"_ref,omitempty"`
	From int64  `json:"from,omitempty"`
	Size int64  `json:"size,omitempty"`
}

type Duplication struct {
	Blocks []*Block `json:"blocks,omitempty"`
}

type Files struct {
	One   *File `json:"1,omitempty"`
	Two   *File `json:"2,omitempty"`
	Three *File `json:"3,omitempty"`
}

type File struct {
	Key         string `json:"key,omitempty"`
	Name        string `json:"name,omitempty"`
	ProjectName string `json:"projectName,omitempty"`
}

type DuplicationsShowOption struct {
	Key  string `url:"key,omitempty"`  // Description:"File key",ExampleValue:"my_project:/src/foo/Bar.php"
	Uuid string `url:"uuid,omitempty"` // Description:"File ID. If provided, 'key' must not be provided.",ExampleValue:"584a89f2-8037-4f7b-b82c-8b45d2d63fb2"
}

// Show Get duplications. Require Browse permission on file's project
func (s *DuplicationsService) Show(opt *DuplicationsShowOption) (v *DuplicationsShowObject, resp *http.Response, err error) {
	err = s.ValidateShowOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "duplications/show", opt)
	if err != nil {
		return
	}
	v = new(DuplicationsShowObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

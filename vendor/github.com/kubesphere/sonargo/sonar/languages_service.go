// Get the list of programming languages supported in this instance.
package sonargo

import "net/http"

type LanguagesService struct {
	client *Client
}

type LanguagesListObject struct {
	Languages []*Language `json:"languages,omitempty"`
}

type Language struct {
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

type LanguagesListOption struct {
	Ps int    `url:"ps,omitempty"` // Description:"The size of the list to return, 0 for all languages",ExampleValue:"25"
	Q  string `url:"q,omitempty"`  // Description:"A pattern to match language keys/names against",ExampleValue:"java"
}

// List List supported programming languages
func (s *LanguagesService) List(opt *LanguagesListOption) (v *LanguagesListObject, resp *http.Response, err error) {
	err = s.ValidateListOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "languages/list", opt)
	if err != nil {
		return
	}
	v = new(LanguagesListObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

// Manage user favorites
package sonargo

import "net/http"

type FavoritesService struct {
	client *Client
}

type FavoritesSearchObject struct {
	Favorites []*Favorite `json:"favorites,omitempty"`
	Paging    Paging      `json:"paging,omitempty"`
}

type Favorite struct {
	Key          string `json:"key,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Qualifier    string `json:"qualifier,omitempty"`
}

type FavoritesAddOption struct {
	Component string `url:"component,omitempty"` // Description:"Component key",ExampleValue:"my_project:/src/foo/Bar.php"
}

// Add Add a component (project, directory, file etc.) as favorite for the authenticated user.<br>Requires authentication and the following permission: 'Browse' on the project of the specified component.
func (s *FavoritesService) Add(opt *FavoritesAddOption) (resp *http.Response, err error) {
	err = s.ValidateAddOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "favorites/add", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type FavoritesRemoveOption struct {
	Component string `url:"component,omitempty"` // Description:"Component key",ExampleValue:"my_project"
}

// Remove Remove a component (project, directory, file etc.) as favorite for the authenticated user.<br>Requires authentication.
func (s *FavoritesService) Remove(opt *FavoritesRemoveOption) (resp *http.Response, err error) {
	err = s.ValidateRemoveOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "favorites/remove", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type FavoritesSearchOption struct {
	P  int `url:"p,omitempty"`  // Description:"1-based page number",ExampleValue:"42"
	Ps int `url:"ps,omitempty"` // Description:"Page size. Must be greater than 0 and less or equal than 500",ExampleValue:"20"
}

// Search Search for the authenticated user favorites.<br>Requires authentication.
func (s *FavoritesService) Search(opt *FavoritesSearchOption) (v *FavoritesSearchObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "favorites/search", opt)
	if err != nil {
		return
	}
	v = new(FavoritesSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

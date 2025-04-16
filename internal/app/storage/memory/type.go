package memory

type userURLs map[string]map[string]struct{}

func (u *userURLs) AddUser(user string) {
	if _, ok := (*u)[user]; !ok {
		(*u)[user] = map[string]struct{}{}
	}
}

func (u *userURLs) AddUserURL(user string, url string) {
	if _, ok := (*u)[user]; ok {
		(*u)[user][url] = struct{}{}
	}
}

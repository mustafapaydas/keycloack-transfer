package users

type User struct {
	Name     *string `json:"name,omitempty"`
	Surname  *string `json:"surname,omitempty"`
	UserName *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
}

type RegisterUser struct {
	Username      *string      `json:"username"`
	Email         *string      `json:"email"`
	FirstName     *string      `json:"firstName"`
	LastName      *string      `json:"lastName"`
	Credentials   []Credential `json:"credentials"`
	EmailVerified bool         `json:"emailVerified"`
	Enabled       bool         `json:"enabled"`
	Attributes    Attribute    `json:"attributes"`
}

type Credential struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type Attribute struct {
	Test string `json:"test"`
}

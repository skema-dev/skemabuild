package auth

type AuthProvider interface {
	StartAuthProcess()
	SaveTokenToFile()
	GetLocalToken() string
	AccessToken() string
}

func NewGithubAuthProvider() AuthProvider {
	return &githubAuth{}
}

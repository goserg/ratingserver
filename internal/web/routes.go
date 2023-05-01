package web

const (
	signin = "/signin"
	signup = "/signup"
	home   = "/"

	api            = "/api"
	apiHome        = api + home
	apiMatchesList = api + "/matches-list"
	apiMatches     = api + "/matches"
	apiGetPlayers  = api + "/players/:id"
)

func Path() map[string]string {
	return map[string]string{
		"SignUp":      signup,
		"SignIn":      signin,
		"Home":        home,
		"Api":         api,
		"ApiHome":     apiHome,
		"ApiNewMatch": apiMatches,
		"ApiMatches":  apiMatchesList,
	}
}

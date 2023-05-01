package web

const (
	signin  = "/signin"
	signup  = "/signup"
	signout = "/signout"
	home    = "/"

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
		"SignOut":     signout,
		"Home":        home,
		"Api":         api,
		"ApiHome":     apiHome,
		"ApiNewMatch": apiMatches,
		"ApiMatches":  apiMatchesList,
	}
}

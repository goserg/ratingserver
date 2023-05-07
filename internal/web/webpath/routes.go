package webpath

const (
	Signin  = "/signin"
	Signup  = "/signup"
	Signout = "/signout"
	Home    = "/"

	Api            = "/api"
	ApiHome        = Api + Home
	ApiMatchesList = Api + "/matches-list"
	ApiNewMatch    = Api + "/matches"
	ApiGetPlayers  = Api + "/players/:id"
	ApiNewPlayer   = Api + "/players"
)

func Path() map[string]string {
	return map[string]string{
		"SignUp":       Signup,
		"SignIn":       Signin,
		"SignOut":      Signout,
		"Home":         Home,
		"Api":          Api,
		"ApiHome":      ApiHome,
		"ApiNewMatch":  ApiNewMatch,
		"ApiMatches":   ApiMatchesList,
		"ApiNewPlayer": ApiNewPlayer,
	}
}

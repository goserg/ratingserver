package web

const (
	signin = "/signin"
	signup = "/signup"
	home   = "/"

	api           = "/api"
	apiHome       = api + home
	apiMatches    = api + "/matches"
	apiGetPlayers = api + "/players/:id"
)

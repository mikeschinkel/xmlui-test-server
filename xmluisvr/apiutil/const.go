package apiutil

import (
	"fmt"
)

const (
	AllowUntrustedQueriesErrorDetail                   = "Your server configuration current restricts access to the API endpoint %s"
	AllowUntrustedQueriesErrorSuggestion               = "Restart your server with the --%s flag to allow raw SQL queries to be used by the %s API endpoint."
	EnsureYourHTTPRequestBodyContainsAValidDBQueryJSON = `Ensure your HTTP request body contains the correct JSON format e.g. {"db_query": "<query_goes_here>", "parameters":[ "<string_value>", <integer_value>, <etc>] where the angle brackets indicate a <placeholder>.`
	EnsureYourDBQueryIsValidForDB                      = `Ensure your database query uses valid syntax for the %s database.`
	EnsureURLBeginsWithPrefix                          = `Ensure URL begins with '/prefix/' and contain a full URL minus the 'http(s)://' protocol. e.g. /proxy/www.microsoft.com/about`
	UnexpectedErrorMatchingURLFileOnGithub             = "Unexpected error while attempting to match requested URL. Check server logs if you have access. Please %s if this issue persists"
	TryRestartingTheServerOrFileOnGithub               = "Try restarting the server. If that does not fix it then please %s"
)

var ReportOnGithubMessageFunc = func() string {
	return fmt.Sprintf("file an issue or a pull request (PR) for us to address this at %s", GitHubRepoURL())
}

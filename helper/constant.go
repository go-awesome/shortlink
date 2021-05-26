package helper

const (
	Production      = 2 // Please set to 1 if in production.
	Domain          = "https://lin.ks/"
	CookieName      = "lin.ks"
	NodeID          = "N1|"                              // Increase per node by value as "N2|", "N3|"... for multiple node
	DBFolder        = "/home/ubuntu/go/src/shortlink/db" // Without trailing slash at the end.
	AddFromToken    = 3                                  // firt N character to get from token and use it in ShortID
	ShortIDToken    = 7                                  // Further added from 1st N char of AddFromToken+NodeID: total=12
	APITokenLength  = 32
	BypassLockGuard = false // set to true if DB is read from multiple instance.
	DB101           = "Failed to Load Read-Heavy database, Please try again!"
	DB102           = "Failed to Load Write-Heavy database, Please try again!"
	CO101           = "Error CO101: Something went wrong! Please try again!"
	ID101           = "Error ID101: Can not create"
	ID102           = "Error ID102: Can not create due to Numbher issue"
	ID103           = "Error ID103: Authorization Token Missing"
	ID104           = "Error ID104: Failed to generate Short ID, Please try again!"
	ID105           = "Error ID105: Can't fetch Short URL"
	ID106           = "Error ID106: Invalid Short URL"
	ID107           = "Error ID107: Failed to Delete, Please try again"
	ID108           = "Error ID108: Invalid Old Long URL"
	ID109           = "Error ID109: Provided Long URL already exists in your Account"
)

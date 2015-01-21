// Command power-league starts a web server that display alternative rankings
// for fantasy sports leagues.
//
// Users grant authentication from the fantasy sports provider (Yahoo) and
// are presented with a list of their current and past leagues. Each league
// contains a table and week-by-week breakdown of how their scores translate
// into the alternative rankings.
//
// To run, this command requires a client key and secret to be passed in at
// runtime. More information about the registration process can be found here:
// http://developer.yahoo.com/fantasysports/guide/GettingStarted.html
//
// Additional configuration options are available. See usage for details.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/Forestmb/goff"
	"github.com/Forestmb/power-league/rankings"
	"github.com/Forestmb/power-league/session"
	"github.com/Forestmb/power-league/site"
	"github.com/golang/glog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

func main() {
	addr := flag.String(
		"address",
		":8080",
		"Address to listen for incoming connections.")
	baseContextFlag := flag.String(
		"baseContext",
		"/power-rankings",
		"Root context of the server.")
	staticFilesLocation := flag.String(
		"static",
		"static",
		"Directory to access static files")
	userCacheDurationSeconds := flag.Int(
		"userCacheDurationSeconds",
		6*60*60,
		"Maximum duration user data will be cached, in seconds. Defaults to"+
			" six hours")
	minimizeAPICalls := flag.Bool(
		"minimizeAPICalls",
		false,
		"Minimize calls to the Yahoo Fantasy Sports API. If enabled, it will "+
			"lower the risk of being throttled but will result in a higher "+
			"average page load time.")
	trackingID := flag.String(
		"trackingID",
		"",
		"Google Analytics tracking ID. If blank, tracking will not be activated")
	clientKey := flag.String(
		"clientKey",
		"",
		"Required client OAuth key. "+
			"See http://developer.yahoo.com/fantasysports/guide/GettingStarted.html"+
			" for more information")
	clientSecret := flag.String(
		"clientSecret",
		"",
		"Required client OAuth secret. "+
			"See http://developer.yahoo.com/fantasysports/guide/GettingStarted.html"+
			" for more information")
	cookieAuthKey := flag.String(
		"cookieAuthKey",
		"",
		"Authentication key for cookie store. "+
			"By default uses a randomly generated key.")
	cookieEncryptionKey := flag.String(
		"cookieEncryptionKey",
		"",
		"Encryption key for cookie store. "+
			"By default uses a randomly generated key.")
	flag.Parse()
	defer glog.Flush()

	invalidInputParameters := false
	if *clientKey == "" {
		fmt.Fprintln(os.Stderr, "power-league: clientKey must be provided")
		invalidInputParameters = true
	}
	if *clientSecret == "" {
		fmt.Fprintln(os.Stderr, "power-league: clientSecret must be provided")
		invalidInputParameters = true
	}
	if invalidInputParameters {
		os.Exit(1)
	}

	// Remove trailing slashes
	baseContext := *baseContextFlag
	size := len(baseContext)
	for size > 0 && baseContext[size-1] == '/' {
		baseContext = baseContext[:size-1]
		size = size - 1
	}
	glog.Infof("starting power rankings site -- context=%s", baseContext)

	rankings.MinimizeAPICalls = *minimizeAPICalls

	// Create cookie store
	var cookieStoreAuthKey []byte
	if len(*cookieAuthKey) == 0 {
		glog.V(2).Infoln("using randomly generated cookie authentication key")
		cookieStoreAuthKey = securecookie.GenerateRandomKey(32)
	} else {
		glog.V(2).Infoln("using cookie authentication key from command line")
		cookieStoreAuthKey = []byte(*cookieAuthKey)
	}

	var cookieStoreEncryptionKey []byte
	if len(*cookieEncryptionKey) == 0 {
		glog.V(2).Infoln("using randomly generated cookie encryption key")
		cookieStoreEncryptionKey = securecookie.GenerateRandomKey(32)
	} else {
		glog.V(2).Infoln("using cookie encryption key from command line")
		cookieStoreEncryptionKey = []byte(*cookieEncryptionKey)
	}

	sessionManager := session.NewManagerWithCache(
		goff.GetConsumer(*clientKey, *clientSecret),
		sessions.NewCookieStore(cookieStoreAuthKey, cookieStoreEncryptionKey),
		*userCacheDurationSeconds)

	site := site.NewSite(
		baseContext, *staticFilesLocation, "templates/html/", *trackingID, sessionManager)
	http.ListenAndServe(*addr, handlers.LoggingHandler(logWriter{}, site.ServeMux))
}

// logWriter implements io.Writer to write to the correct logging
// implementation.
type logWriter struct{}

// Write the input to the correct logging implementation
func (l logWriter) Write(p []byte) (int, error) {
	glog.Infof(string(p))
	return len(p), nil
}

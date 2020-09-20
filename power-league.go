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
		":443",
		"Address to listen for incoming connections.")
	noTLS := flag.Bool("noTLS", false, "Disable TLS.")
	tlsCert := flag.String("tlsCert", "./certs/localhost.crt", "TLS certificate if using HTTPS.")
	tlsKey := flag.String("tlsKey", "./certs/localhost.key", "TLS private key if using HTTPS.")
	baseContextFlag := flag.String(
		"baseContext",
		"/",
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
	totalCacheSize := flag.Int64(
		"totalCacheSize",
		10000,
		"Maximum number of responses that well be cached across all users.")
	minimizeAPICalls := flag.Bool(
		"minimizeAPICalls",
		false,
		"Minimize calls to the Yahoo Fantasy Sports API. If enabled, it will "+
			"lower the risk of being throttled but will result in a higher "+
			"average page load time.")
	trackingID := flag.String(
		"trackingID",
		os.Getenv("GA_TRACKING_ID"),
		"Google Analytics tracking ID. If blank, tracking will not be activated. "+
			"Defaults to value of the GA_TRACKING_ID environment variable.")
	clientKey := flag.String(
		"clientKey",
		"",
		"Required client OAuth key. Defaults to the value of OAUTH_CLIENT_KEY. "+
			"See http://developer.yahoo.com/fantasysports/guide/GettingStarted.html"+
			" for more information")
	clientSecret := flag.String(
		"clientSecret",
		"",
		"Required client OAuth secret. Defaults to the value of OAUTH_CLIENT_SECRET. "+
			"See http://developer.yahoo.com/fantasysports/guide/GettingStarted.html"+
			" for more information")
	clientRedirectURL := flag.String(
		"clientRedirectURL",
		"",
		"Client redirect URL. Defaults to the listening address.")
	cookieAuthKey := flag.String(
		"cookieAuthKey",
		"",
		"Authentication key for cookie store. Defaults to the value of COOKIE_AUTH_KEY. "+
			"By default uses a randomly generated key.")
	cookieEncryptionKey := flag.String(
		"cookieEncryptionKey",
		"",
		"Encryption key for cookie store. Defaults to the value of COOKIE_ENCRYPTION_KEY. "+
			"By default uses a randomly generated key.")
	flag.Parse()
	defer glog.Flush()

	invalidInputParameters := false
	if *clientKey == "" {
		envValue := os.Getenv("OAUTH_CLIENT_KEY")
		clientKey = &envValue
	}
	if *clientKey == "" {
		fmt.Fprintln(os.Stderr, "power-league: clientKey must be provided")
		invalidInputParameters = true
	}

	if *clientSecret == "" {
		envValue := os.Getenv("OAUTH_CLIENT_SECRET")
		clientSecret = &envValue
	}
	if *clientSecret == "" {
		fmt.Fprintln(os.Stderr, "power-league: clientSecret must be provided")
		invalidInputParameters = true
	}

	if !*noTLS {
		if *tlsCert == "" {
			fmt.Fprintln(os.Stderr, "power-league: tlsCert must be provided")
			invalidInputParameters = true
		}
		if *tlsKey == "" {
			fmt.Fprintln(os.Stderr, "power-league: tlsKey must be provided")
			invalidInputParameters = true
		}
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
		envValue := os.Getenv("COOKIE_AUTH_KEY")
		cookieAuthKey = &envValue
	}
	if len(*cookieAuthKey) == 0 {
		glog.V(2).Infoln("using randomly generated cookie authentication key")
		cookieStoreAuthKey = securecookie.GenerateRandomKey(32)
	} else {
		glog.V(2).Infoln("using cookie authentication key from command line")
		cookieStoreAuthKey = []byte(*cookieAuthKey)
	}

	var cookieStoreEncryptionKey []byte
	if len(*cookieEncryptionKey) == 0 {
		envValue := os.Getenv("COOKIE_ENCRYPTION_KEY")
		cookieEncryptionKey = &envValue
	}
	if len(*cookieEncryptionKey) == 0 {
		glog.V(2).Infoln("using randomly generated cookie encryption key")
		cookieStoreEncryptionKey = securecookie.GenerateRandomKey(32)
	} else {
		glog.V(2).Infoln("using cookie encryption key from command line")
		cookieStoreEncryptionKey = []byte(*cookieEncryptionKey)
	}

	authContext := fmt.Sprintf("%s/auth", baseContext)
	sessionManager := session.NewManagerWithCache(
		oauth2ConsumerProvider{
			tls:          !*noTLS,
			clientKey:    *clientKey,
			clientSecret: *clientSecret,
			redirectURL:  *clientRedirectURL,
			authContext:  authContext,
		},
		sessions.NewCookieStore(cookieStoreAuthKey, cookieStoreEncryptionKey),
		*userCacheDurationSeconds,
		*totalCacheSize)

	site := site.NewSite(
		!*noTLS, baseContext, *staticFilesLocation, "templates/html/", *trackingID, sessionManager)
	var err error
	if *noTLS {
		err = http.ListenAndServe(*addr, handlers.LoggingHandler(logWriter{}, site.ServeMux))
	} else {
		err = http.ListenAndServeTLS(*addr, *tlsCert, *tlsKey, handlers.LoggingHandler(logWriter{}, site.ServeMux))
	}
	if err != nil {
		glog.Exit("ListenAndServe: ", err)
	}
}

// logWriter implements io.Writer to write to the correct logging
// implementation.
type logWriter struct{}

// Write the input to the correct logging implementation
func (l logWriter) Write(p []byte) (int, error) {
	glog.Infof(string(p))
	return len(p), nil
}

// oauth2ConsumerProvider implements session.Consumer to provide the OAuth 2
// config for this application
type oauth2ConsumerProvider struct {
	tls          bool
	clientKey    string
	clientSecret string
	redirectURL  string
	authContext  string
}

func (o oauth2ConsumerProvider) Get(r *http.Request) session.Consumer {
	redirectURL := o.redirectURL
	if redirectURL == "" {
		protocol := "https"
		if !o.tls {
			protocol = "http"
		}
		redirectURL = fmt.Sprintf("%s://%s%s", protocol, r.Host, o.authContext)
	}
	return goff.GetOAuth2Config(o.clientKey, o.clientSecret, redirectURL)
}

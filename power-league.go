package main

import (
	"flag"
	"net/http"

	"github.com/Forestmb/goff"
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
	clientKey := flag.String(
		"clientKey",
		"",
		"Client OAuth key. "+
			"See http://developer.yahoo.com/fantasysports/guide/GettingStarted.html"+
			" for more information")
	clientSecret := flag.String(
		"clientSecret",
		"",
		"Client OAuth secret. "+
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

	// Remove trailing slashes
	baseContext := *baseContextFlag
	size := len(baseContext)
	for size > 0 && baseContext[size-1] == '/' {
		baseContext = baseContext[:size-1]
		size = size - 1
	}
	glog.Infof("starting power rankings site -- context=%s", baseContext)

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

	sessionManager := session.NewSessionManager(
		goff.GetConsumer(*clientKey, *clientSecret),
		sessions.NewCookieStore(cookieStoreAuthKey, cookieStoreEncryptionKey))

	site := site.NewSite(
		baseContext, *staticFilesLocation, "templates/html/", sessionManager)
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

# Power League [![GoDoc](https://godoc.org/github.com/Forestmb/power-league?status.png)](https://godoc.org/github.com/Forestmb/power-league) #

Power League is a web application that calculates alternative rankings for
Yahoo Fantasy Sports leagues.

This application is written using the Go programming language and is licensed
under the [New BSD license](
https://github.com/Forestmb/power-league/blob/master/LICENSE).

## Rnning ##

The `power-league` application is distributed as a [docker container](
https://github.com/users/Forestmb/packages/container/package/power-league) and
can be run with any compatible container runtime.

To run the server, first obtain a [Yahoo Fantasy Sports client key and secret](
http://developer.yahoo.com/fantasysports/guide/GettingStarted.html). When registering you
will be asked to enter in a redirect URL, which should be in the format
`https://<hostname>/<base-context>/auth'.

For running on your local machine you will likely need to use a redirect URL of
`https://127.0.0.1/auth`. As of September 2020 you cannot register an
application with any of the following in the redirect URL:

- An `http://` protocol
- A `localhost` hostname
- A non-default port (not 443)

Once your application is registered you can pass the client key/secret into the
application by setting the environment variables `OAUTH_CLIENT_KEY` and
`OAUTH_CLIENT_SECRET` and passing them when running docker:

    $ docker run --rm \
             -p443:443 \
             -e OAUTH_CLIENT_KEY \
             -e OAUTH_CLIENT_SECRET \
             ghcr.io/forestmb/power-league:latest

After it has started, browse to `https://127.0.0.1/` to access the application..
By default a self-signed test certificate will be used. To provide your own TLS
certificates you can mount them using docker volumes and with additional
command-line flags:

    $ docker run --rm \
             -p443:443 \
             -e OAUTH_CLIENT_KEY \
             -e OAUTH_CLIENT_SECRET \
             -v /path/to/cert.pem:/app/certs/cert.pem \
             -v /path/to/key.pem:/app/certs/key.pem \
             ghcr.io/forestmb/power-league:latest \
             -tlsCert /app/certs/cert.pem \
             -tlsKey /app/certs/key.pem

If you are running the application behind a reverse proxy like nginx or Apache
and do not need or wish to run with HTTPS, you can opt out when running the
server. However when registering your application you will still need to provide
an HTTPS redirect URL, and you should pass in the publicly-accessible redirect
URL when starting the container:

    $ docker run --rm \
             -p8080:8080 \
             -e OAUTH_CLIENT_KEY \
             -e OAUTH_CLIENT_SECRET \
             ghcr.io/forestmb/power-league:latest \
             -noTLS \
             -address :8008 \
             -redirectURL https://public.url.example.com/auth

Once your server is running, visit it in the web browser and sign in. After
granting access you can view the rankings for any of your current or past
leagues:

![Example Screenshot](https://raw.github.com/Forestmb/power-league/master/doc/screenshots/rankings.png)

## Building ##

Building requires an installation of the [Go programming language tools](
https://golang.org/doc/install) and/or a [Docker](https://www.docker.com/)
compatible container build engine. To build:

    # Build using docker
    $ docker build -t power-league:latest .

    # Build server locally
    $ ./build.sh

To run the local build before every commit, use:

    $ ln -s "$(pwd)/build.sh" .git/hooks/pre-commit

## Options ##

Command line flags can be passed when running either locally or as additional docker run
command arguments.

    Usage of ./power-league:
      -address string
        	Address to listen for incoming connections. (default ":443")
      -alsologtostderr
        	log to standard error as well as files
      -baseContext string
        	Root context of the server. (default "/")
      -clientKey string
        	Required client OAuth key. Defaults to the value of OAUTH_CLIENT_KEY.
            See http://developer.yahoo.com/fantasysports/guide/GettingStarted.html
            for more information
      -clientSecret string
        	Required client OAuth secret. Defaults to the value of
            OAUTH_CLIENT_SECRET. See http://developer.yahoo.com/fantasysports/guide/GettingStarted.html
            for more information
      -cookieAuthKey string
        	Authentication key for cookie store. Defaults to the value of
            COOKIE_AUTH_KEY. By default uses a randomly generated key.
      -cookieEncryptionKey string
        	Encryption key for cookie store. Defaults to the value of
            COOKIE_ENCRYPTION_KEY. By default uses a randomly generated key.
      -log_backtrace_at value
        	when logging hits line file:N, emit a stack trace
      -log_dir string
        	If non-empty, write log files in this directory
      -logtostderr
        	log to standard error instead of files
      -minimizeAPICalls
        	Minimize calls to the Yahoo Fantasy Sports API. If enabled, it will
            lower the risk of being throttled but will result in a higher
            average page load time.
      -noTLS
        	Disable TLS.
      -static string
        	Directory to access static files (default "static")
      -stderrthreshold value
        	logs at or above this threshold go to stderr
      -tlsCert string
        	TLS certificate if using HTTPS. (default "./certs/localhost.crt")
      -tlsKey string
        	TLS private key if using HTTPS. (default "./certs/localhost.key")
      -totalCacheSize int
        	Maximum number of responses that well be cached across all users.
            (default 10000)
      -trackingID string
        	Google Analytics tracking ID. If blank, tracking will not be
            activated. Defaults to value of the GA_TRACKING_ID environment
            variable.
      -userCacheDurationSeconds int
        	Maximum duration user data will be cached, in seconds. Defaults to
            six hours (default 21600)
      -v value
        	log level for V logs
      -vmodule value
        	comma-separated list of pattern=N settings for file-filtered logging

# Power League [![GoDoc](https://godoc.org/github.com/Forestmb/power-league?status.png)](https://godoc.org/github.com/Forestmb/power-league) [![Build Status](https://travis-ci.org/Forestmb/power-league.png?branch=master)](https://travis-ci.org/Forestmb/power-league) [![Coverage Status](https://coveralls.io/repos/Forestmb/power-league/badge.png?branch=master)](https://coveralls.io/r/Forestmb/power-league?branch=master) #

Power League is a web application that calculates alternative rankings for
Yahoo Fantasy Sports leagues.

This application is written using the Go programming language and is licensed
under the [New BSD license](
https://github.com/Forestmb/power-league/blob/master/LICENSE).

## Building ##

Building requires an installation of the [Go programming language tools](
https://golang.org/doc/install). Once installed, you can follow [this guide](
https://golang.org/doc/code.html) to become familiar with the structure of
most Go projects.

Once your `GOPATH` is set up, you can build this project with the following:

    $ go get github.com/Forestmb/power-league
    $ cd $GOPATH/src/github.com/Forestmb/power-league
    $ ./build.sh

To make sure this build runs before every commit, use:

    $ ln -s "$(pwd)/build.sh" .git/hooks/pre-commit

## Running ##

To run a built instance locally, you must first create a configuration file.

    $ cp server.conf.template server.conf

Next, obtain a [Yahoo Fantasy Sports client key and secret](
http://developer.yahoo.com/fantasysports/guide/GettingStarted.html) and copy
the values into `server.conf`

Then, run the application

    $ ./server.sh start

The application can then be accessed at `http://localhost:8080/power-rankings`.
Once signed in, you can view the rankings for any of your current or past leagues:

![Example Screenshot](https://raw.github.com/Forestmb/power-league/master/doc/screenshots/rankings.png)

## Deploying ##

If you wish to deploy the application to a remote server, you can use the
`package.sh` utility by passing in the name of the host like so:

    $ ./package.sh -D <host>

This packages a built application, copies it to the remote host, stops
the existing instance if necessary, and starts the application. By default it
uses the `server.conf` file to configure the application, but you can override
it by defining a host-specific file named `server.conf.<host>`.

The `deploydir` variable in the configuration file determines where on the
remote host the instances of the application are kept. The deploy utility
maintains the last two deployed instances of the application in `deploydir`
(`current` and `previous`), and archives old versions in the
`<deploydir>/old/` directory.

## Options ##

Command line flags can be passed when running `server.sh` or by appending them
to the `server_args` variable in `server.conf`.

    Usage of ./power-league:
      -address=":8080": Address to listen for incoming connections.
      -alsologtostderr=false: log to standard error as well as files
      -baseContext="/power-rankings": Root context of the server.
      -clientKey="": Required client OAuth key. See http://developer.yahoo.com/fantasysports/guide/GettingStarted.html for more information
      -clientSecret="": Required client OAuth secret. See http://developer.yahoo.com/fantasysports/guide/GettingStarted.html for more information
      -cookieAuthKey="": Authentication key for cookie store. By default uses a randomly generated key.
      -cookieEncryptionKey="": Encryption key for cookie store. By default uses a randomly generated key.
      -log_backtrace_at=:0: when logging hits line file:N, emit a stack trace
      -log_dir="": If non-empty, write log files in this directory
      -logtostderr=false: log to standard error instead of files
      -static="static": Directory to access static files
      -stderrthreshold=0: logs at or above this threshold go to stderr
      -trackingID="": Google Analytics tracking ID. If blank, tracking will not be activated
      -userCacheDurationSeconds=21600: Maximum duration user data will be cached, in seconds. Defaults to six hours
      -v=0: log level for V logs
      -vmodule=: comma-separated list of pattern=N settings for file-filtered logging

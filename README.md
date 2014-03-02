# Power League [![GoDoc](https://godoc.org/github.com/Forestmb/power-league?status.png)](https://godoc.org/github.com/Forestmb/power-league) [![Build Status](https://travis-ci.org/Forestmb/power-league.png?branch=master)](https://travis-ci.org/Forestmb/power-league) #

Power League is a web application that calculates alternative rankings for
Yahoo Fantasy Sports leagues.

This application is written using the Go programming language and is licensed
under the [New BSD license](
https://github.com/Forestmb/power-league/blob/master/LICENSE).

## Building ##

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
      -v=0: log level for V logs
      -vmodule=: comma-separated list of pattern=N settings for file-filtered logging

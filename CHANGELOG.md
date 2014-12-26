# Changelog #

## 0.3.0 (TBD) ##

- Traffic analysis through Google Analytics.
- Added user preference for displaying all-play records instead of power scores.
- Added column sorting to rankings page.
- Added league rank and record to rankings page.
- Overall rankings can now be exported to CSV or a newsletter template.
- Changed signature of `rankings.GetPowerData`.
- Renamed `session.SessionManager` to `session.Manager` and updated associated
  function names.
- Replaced all occurrences of `rankings.Record` with the equivalent type
  `goff.Record`.

## 0.2.0 (2014-09-05) ##

- Power ranking projections for unfinished leagues.
- Caches fantasy content for each user for up to 6 hours (configurable).
- Fixed error message when displaying leagues pre-draft.

## 0.1.0 (2014-03-02) ##

Initial public release

- OAuth 1.0 for authentication.
- Support for fantasy football leagues.

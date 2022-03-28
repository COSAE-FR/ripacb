# RIP ACB Server and client

## Server (acbd)
This is an implementation of the pfSense [Automatic Config Backup](https://docs.netgate.com/pfsense/en/latest/backup/autoconfigbackup.html) server.

The `acbd` service can run multiple ACB servers with different sets of features enabled.

### Configuration

#### Global settings

- `store`: type of data storage, only `directory` (default) available
- `path`: path to the store's directory (required)
- `logging`: log settings:
  - `file`: log file (`stderr` if empty)
  - `level`: log level (one of trace,debug,info,warning,error, info if empty)
  - `max_size`: maximum size of the log file
  - `max_age`: maximum age of the logs before rotating
  - `max_backups`: how many rotated log file to keep
- `servers`: list of server configurations, see below

#### Server settings

- `ip`: IP address to listen on (default: `0.0.0.0`)
- `port`: TCP port to listen on (default: `80`)
- `rate`: number of requests per minute, per client IP to accept (rate limiting, default `60`)
- `features`: enabled features:
  - `allow_save`: allow saving new configuration revision
  - `allow_new`: allow saving new configuration revision for new devices
  - `allow_restore_user`: allow restoring user saved revisions
  - `max_backups`: maximum revision count to keep
  - `is_portal`: this server handles only portal requests
- `tls`: TLS configuration:
  - `hostname`: server hostname (required if TLS is enabled)
  - `cert`: server certificate or path to the server certificate file or `acme` if autocert must be enabled
  - `key`: server key or path to the server key file (required if `cert` is not `acme`)
  - `ca`: CA chain or path to the CA chain file
  - `cache`: cache directory path for the ACME service
  - `client_ca`: CA or path to the CA file for validating client certificates

## Client (ripacb)

This is a TUI client for the ACB server.



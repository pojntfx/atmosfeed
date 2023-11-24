# Atmosfeed

![Logo](./docs/logo-readme.png)

Create custom Bluesky feeds with WebAssembly and Scale.

## Overview

🚧 This project is a work-in-progress! Instructions will be added as soon as it is usable. 🚧

## Installation

Atmosfeed is available to the public and can be used by opening it in a browser:

<a href="https://atmosfeed.p8.lu/"><img src="https://github.com/pojntfx/webnetesctl/raw/main/img/launch.png" alt="PWA badge" width="200"/></a>

If you prefer to self-host, see [contributing](#contributing); static binaries for the manager and worker, a `.tar.gz` archive for the frontend and an OCI image for containerization are also available on [GitHub releases](https://github.com/pojntfx/atmosfeed/releases) and [GitHub container registry](https://github.com/pojntfx/atmosfeed/packages) respectively.

In addition to this publicly hosted infrastructure, a client CLI is also provided in the form of static binaries in order to make feed deployment and management from your local system easier.

On Linux, you can install them like so:

```shell
curl -L -o /tmp/atmosfeed-client "https://github.com/pojntfx/atmosfeed/releases/latest/download/atmosfeed-client.linux-$(uname -m)"
sudo install /tmp/atmosfeed-client /usr/local/bin
```

On macOS, you can use the following:

```shell
curl -L -o /tmp/atmosfeed-client "https://github.com/pojntfx/atmosfeed/releases/latest/download/atmosfeed-client.darwin-$(uname -m)"
sudo install /tmp/atmosfeed-client /usr/local/bin
```

On Windows, the following should work (using PowerShell as administrator):

```shell
Invoke-WebRequest https://github.com/pojntfx/atmosfeed/releases/latest/download/atmosfeed-client.windows-x86_64.exe -OutFile \Windows\System32\atmosfeed-client.exe
```

You can find binaries for more operating systems and architectures on [GitHub releases](https://github.com/pojntfx/atmosfeed/releases).

## Reference

### Command Line Arguments

#### Server CLI

```shell
$ atmosfeed-server --help
Create custom Bluesky feeds with WebAssembly and Scale.
Find more information at:
https://github.com/pojntfx/atmosfeed

Usage:
  atmosfeed-server [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  manager     Start an Atmosfeed manager
  worker      Start an Atmosfeed worker

Flags:
  -h, --help                  help for atmosfeed-server
      --postgres-url string   PostgreSQL URL (default "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable")
      --redis-url string      Redis URL (default "redis://localhost:6379/0")
      --s3-url string         S3 URL (default "http://minioadmin:minioadmin@localhost:9000?bucket=atmosfeed")
      --verbose               Whether to enable verbose logging

Use "atmosfeed-server [command] --help" for more information about a command.
```

<details>
  <summary>Expand subcommand reference</summary>

##### Manager

```shell
$ atmosfeed-server manager --help
Start an Atmosfeed manager

Usage:
  atmosfeed-server manager [flags]

Aliases:
  manager, m

Flags:
      --bgs-url string              BGS URL (default "https://bsky.network")
      --delete-all-posts            Whether to delete all posts from the index on startup (required for compliance with the EU right to be forgotten/GDPR article 17; deletions during uptime are handled using delete commits) (default true)
      --feed-generator-did string   DID of the feed generator (typically the hostname of the publicly reachable URL) (default "did:web:manager.atmosfeed.p8.lu")
      --feed-generator-url string   Publicly reachable URL of the feed generator (default "https://manager.atmosfeed.p8.lu")
  -h, --help                        help for manager
      --laddr string                Listen address (default ":1337")
      --limit int                   Maximum amount of posts to return for a feed (default 100)
      --origin string               Allowed CORS origin (default "https://atmosfeed.p8.lu")
      --ttl duration                Maximum age of posts to return for a feed (default 6h0m0s)

Global Flags:
      --postgres-url string   PostgreSQL URL (default "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable")
      --redis-url string      Redis URL (default "redis://localhost:6379/0")
      --s3-url string         S3 URL (default "http://minioadmin:minioadmin@localhost:9000?bucket=atmosfeed")
      --verbose               Whether to enable verbose logging
```

##### Worker

```shell
$ atmosfeed-server worker --help
Start an Atmosfeed worker

Usage:
  atmosfeed-server worker [flags]

Aliases:
  worker, w

Flags:
      --classifier-timeout duration   Amount of time after which to stop a classifier Scale function from running (default 1s)
  -h, --help                          help for worker
      --working-directory string      Working directory to use (default "/home/pojntfx/.local/share/atmosfeed/var/lib/atmosfeed")

Global Flags:
      --postgres-url string   PostgreSQL URL (default "postgresql://postgres@localhost:5432/atmosfeed?sslmode=disable")
      --redis-url string      Redis URL (default "redis://localhost:6379/0")
      --s3-url string         S3 URL (default "http://minioadmin:minioadmin@localhost:9000?bucket=atmosfeed")
      --verbose               Whether to enable verbose logging
```

</details>

#### Client CLI

```shell
$ atmosfeed-client --help
Create custom Bluesky feeds with WebAssembly and Scale.
Find more information at:
https://github.com/pojntfx/atmosfeed

Usage:
  atmosfeed-client [command]

Available Commands:
  apply           Create or update a feed on an Atmosfeed server
  completion      Generate the autocompletion script for the specified shell
  delete          Delete a feed from an Atmosfeed server
  delete-userdata Delete all user data from an Atmosfeed server
  dev             Develop a feed classifier locally
  export-userdata Export all user data from an Atmosfeed server
  help            Help about any command
  list            List published feeds on an Atmosfeed server
  publish         Publish a feed to a Bluesky PDS
  resolve         Resolve a handle to a DID
  unpublish       Unpublish a feed from a Bluesky PDS

Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
  -h, --help                   help for atmosfeed-client
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")

Use "atmosfeed-client [command] --help" for more information about a command.
```

<details>
  <summary>Expand subcommand reference</summary>

##### Apply

```shell
$ atmosfeed-client apply --help
Create or update a feed on an Atmosfeed server

Usage:
  atmosfeed-client apply [flags]

Aliases:
  apply, a

Flags:
      --clear-pinned              Whether to clear the pinned post field
      --feed-classifier string    Path to the feed classifier to upload (default "local-trending-latest.scale")
      --feed-rkey string          Machine-readable key for the feed (default "trending")
  -h, --help                      help for apply
      --pinned-feed-did string    DID of the pinned post for the feed (if left empty, no post will be pinned; empty values don't overwrite non-empty values, see --clear-pinned)
      --pinned-feed-rkey string   Machine-readable key of the pinned post for the feed (if left empty, no post will be pinned; empty values don't overwrite non-empty values, see --clear-pinned)

Global Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")
```

##### Publish

```shell
$ atmosfeed-client publish --help
Publish a feed to a Bluesky PDS

Usage:
  atmosfeed-client publish [flags]

Aliases:
  publish, p

Flags:
      --feed-description string     Description for the feed (default "An example trending feed for Atmosfeed")
      --feed-generator-did string   DID of the feed generator (typically the hostname of the publicly reachable URL) (default "did:web:manager.atmosfeed.p8.lu")
      --feed-name string            Human-readable name for the feed) (default "Atmosfeed Trending")
      --feed-rkey string            Machine-readable key for the feed (default "trending")
  -h, --help                        help for publish

Global Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")
```

##### List

```shell
$ atmosfeed-client list --help
List published feeds on an Atmosfeed server

Usage:
  atmosfeed-client list [flags]

Aliases:
  list, l

Flags:
  -h, --help   help for list

Global Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")
```

##### Unpublish

```shell
$ atmosfeed-client unpublish --help
Unpublish a feed from a Bluesky PDS

Usage:
  atmosfeed-client unpublish [flags]

Aliases:
  unpublish, u

Flags:
      --feed-rkey string   Machine-readable key for the feed (default "trending")
  -h, --help               help for unpublish

Global Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")
```

##### Delete

```shell
$ atmosfeed-client delete --help
Delete a feed from an Atmosfeed server

Usage:
  atmosfeed-client delete [flags]

Aliases:
  delete, d

Flags:
      --feed-rkey string   Machine-readable key for the feed (default "trending")
  -h, --help               help for delete

Global Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")
```

##### Dev

```shell
$ atmosfeed-client dev --help
Develop a feed classifier locally

Usage:
  atmosfeed-client dev [flags]

Aliases:
  dev, v

Flags:
      --bgs-url string           BGS URL (default "https://bsky.network")
      --feed-classifier string   Path to the feed classifier to test (default "local-trending-latest.scale")
      --frontend-url string      Bluesky frontend URL to use when logging posts (default "https://bsky.app")
  -h, --help                     help for dev
      --max-posts int            Maximum amount of posts to store in memory before clearing the cache (default 1048576)
      --min-weight int           Minimum weight value the classifier has to return for a post to log it
      --quiet                    Whether to silently ignore any non-fatal errors (default true)
      --verbose                  Whether to enable verbose logging

Global Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")
```

##### Resolve

```shell
$ atmosfeed-client resolve --help
Resolve a handle to a DID

Usage:
  atmosfeed-client resolve [flags]

Aliases:
  resolve, p

Flags:
      --handle string   Handle/username/domain to resolve
  -h, --help            help for resolve

Global Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")
```

##### Export User Data

```shell
$ atmosfeed-client export-userdata --help
Export all user data from an Atmosfeed server

Usage:
  atmosfeed-client export-userdata [flags]

Aliases:
  export-userdata, eu

Flags:
  -h, --help         help for export-userdata
      --out string   Directory to export user data to (default "atmosfeed-userdata")

Global Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")
```

##### Delete User Data

```shell
$ atmosfeed-client delete-userdata --help
Delete all user data from an Atmosfeed server

Usage:
  atmosfeed-client delete-userdata [flags]

Aliases:
  delete-userdata, du

Flags:
  -h, --help   help for delete-userdata

Global Flags:
      --atmosfeed-url string   Atmosfeed server URL (default "https://manager.atmosfeed.p8.lu")
      --password string        Bluesky password, preferably an app password (get one from https://bsky.app/settings/app-passwords)
      --pds-url string         PDS URL (default "https://bsky.social")
      --username string        Bluesky username (default "example.bsky.social")
```

</details>

### Environment Variables

All command line arguments described above can also be set using environment variables; for example, to set `--pds-url` to `https://bsky.social` with an environment variable, use `ATMOSFEED_PDS_URL='https://bsky.social'`. In addition to this, there are also some aliases in place for compatibility with Railway/Heroku conventions; for example, you can `export PORT=3000` to listen on port 3000 as an alternative to setting the listen address, or `export DATABASE_URL='postgresql://username:password@myhost:5432/mydb'` to set the database URL. For more info, see the [command line arguments](#command-line-arguments).

## Acknowledgements

- [loopholelabs/scale](https://github.com/loopholelabs/scale) provides the WebAssembly-based plugin system.
- [sqlc-dev/sqlc](https://github.com/sqlc-dev/sqlc) provides the SQL library.
- [pressly/goose](https://github.com/pressly/goose) provides migration support.
- [bluesky-social/indigo](https://github.com/bluesky-social/indigo) provides the Bluesky API client.

## Contributing

To contribute, please use the [GitHub flow](https://guides.github.com/introduction/flow/) and follow our [Code of Conduct](./CODE_OF_CONDUCT.md).

To build and start a development version of Atmosfeed locally, run the following:

```shell
# Download the source code
git clone https://github.com/pojntfx/atmosfeed.git
cd atmosfeed

# Setup dependencies
docker rm -f atmosfeed-postgres && docker run -d --name atmosfeed-postgres -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=atmosfeed postgres
docker rm -f atmosfeed-redis && docker run --name atmosfeed-redis -p 6379:6379 -d redis
docker rm -f atmosfeed-minio && docker run --name atmosfeed-minio -p 9000:9000 -d minio/minio server /data

make -j$(nproc) depend

# Start manager
export ATMOSFEED_ORIGIN='http://localhost:3000'
export ATMOSFEED_FEED_GENERATOR_DID='did:web:atmosfeed.serveo.net'
make -j$(nproc) depend/cli && go run ./cmd/atmosfeed-server manager

# Start a tunnel to reach the manager from the public internet
ssh -R atmosfeed.serveo.net:80:localhost:1337 serveo.net

# Start worker(s)
make -j$(nproc) depend/cli && go run ./cmd/atmosfeed-server worker --working-directory ~/.local/share/atmosfeed/var/lib/atmosfeed/worker-1
make -j$(nproc) depend/cli && go run ./cmd/atmosfeed-server worker --working-directory ~/.local/share/atmosfeed/var/lib/atmosfeed/worker-2

# Download example feeds
curl -Lo out/local-everything-latest.scale https://github.com/pojntfx/bluesky-feeds/releases/download/release-main/local-everything-latest.scale
curl -Lo out/local-german-latest.scale https://github.com/pojntfx/bluesky-feeds/releases/download/release-main/local-german-latest.scale
curl -Lo out/local-question-latest.scale https://github.com/pojntfx/bluesky-feeds/releases/download/release-main/local-question-latest.scale
curl -Lo out/local-trending-latest.scale https://github.com/pojntfx/bluesky-feeds/releases/download/release-main/local-trending-latest.scale

# Deploy example feeds
export ATMOSFEED_PASSWORD='asdf'
export ATMOSFEED_USERNAME='pojntfxtesting.bsky.social'
export ATMOSFEED_ATMOSFEED_URL='http://localhost:1337'
export ATMOSFEED_FEED_GENERATOR_DID='did:web:atmosfeed.serveo.net'

go run ./cmd/atmosfeed-client/ apply --feed-rkey everything --feed-classifier ./out/local-everything-latest.scale
go run ./cmd/atmosfeed-client/ publish --feed-rkey everything --feed-name 'Atmosfeed Everything' --feed-description 'Newest posts on Bluesky (testing feed)'

go run ./cmd/atmosfeed-client/ apply --feed-rkey questions --feed-classifier ./out/local-questions-latest.scale
go run ./cmd/atmosfeed-client/ publish --feed-rkey questions --feed-name 'Atmosfeed Questions' --feed-description 'Most popular questions on Bluesky in the last 24h (testing feed).'

go run ./cmd/atmosfeed-client/ apply --feed-rkey german --feed-classifier ./out/local-german-latest.scale
go run ./cmd/atmosfeed-client/ publish --feed-rkey german --feed-name 'Atmosfeed German' --feed-description 'Most popular German posts on Bluesky in the last 24h (testing feed)'

go run ./cmd/atmosfeed-client/ apply --feed-rkey trending --feed-classifier ./out/local-trending-latest.scale
go run ./cmd/atmosfeed-client/ publish --feed-rkey trending --feed-name 'Atmosfeed Trending' --feed-description 'Most popular trending posts on Bluesky in the last 24h (testing feed)'

# Remove example feeds
go run ./cmd/atmosfeed-client/ delete --feed-rkey questions
go run ./cmd/atmosfeed-client/ unpublish --feed-rkey questions

go run ./cmd/atmosfeed-client/ delete --feed-rkey german
go run ./cmd/atmosfeed-client/ unpublish --feed-rkey german

go run ./cmd/atmosfeed-client/ delete --feed-rkey everything
go run ./cmd/atmosfeed-client/ unpublish --feed-rkey everything

go run ./cmd/atmosfeed-client/ delete --feed-rkey trending
go run ./cmd/atmosfeed-client/ unpublish --feed-rkey trending

# Start frontend
cd frontend
bun dev # Now visit http://localhost:3000 to open the frontend and sign in

# Export or delete user data for privacy & interoperability
go run ./cmd/atmosfeed-client/ export-userdata --out ./out/atmosfeed-userdata
go run ./cmd/atmosfeed-client/ delete-userdata
```

Have any questions or need help? Chat with us [on Matrix](https://matrix.to/#/#skysweeper:matrix.org?via=matrix.org)!

## License

Atmosfeed (c) 2023 Felicitas Pojtinger and contributors

SPDX-License-Identifier: Apache-2.0

# Atmosfeed

![Logo](./docs/logo-readme.png)

Create custom Bluesky feeds with WebAssembly and Scale.

## Overview

🚧 This project is a work-in-progress! Instructions will be added as soon as it is usable. 🚧

## Installation

Atmosfeed is available to the public and can be used by opening it in a browser:

<a href="https://atmosfeed.p8.lu/"><img src="https://github.com/pojntfx/webnetesctl/raw/main/img/launch.png" alt="PWA badge" width="200"/></a>

If you prefer to self-host, see [contributing](#contributing); static binaries for the manager and worker, a `.tar.gz` archive for the frontend and an OCI image for containerization are also available on [GitHub releases](https://github.com/pojntfx/atmosfeed/releases) and [GitHub container registry](https://github.com/pojntfx/atmosfeed/packages) respectively.

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

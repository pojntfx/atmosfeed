# Atmosfeed Demo

```shell
# Dependencies
git clone https://github.com/pojntfx/atmosfeed.git
cd atmosfeed

docker rm -f atmosfeed-postgres && docker run -d --name atmosfeed-postgres -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=atmosfeed postgres
docker rm -f atmosfeed-redis && docker run --name atmosfeed-redis -p 6379:6379 -d redis
docker rm -f atmosfeed-minio && docker run --name atmosfeed-minio -p 9000:9000 -d minio/minio server /data

make -j$(nproc) depend
```

```shell
# Manager
export ATMOSFEED_ORIGIN='http://localhost:3000'
make -j$(nproc) depend/cli && go run ./cmd/atmosfeed-server manager
```

```shell
# Tunnel
ssh -R atmosfeed.serveo.net:80:localhost:1337 serveo.net
```

```shell
# Workers
make -j$(nproc) depend/cli && go run ./cmd/atmosfeed-server worker --working-directory ~/.local/share/atmosfeed/var/lib/atmosfeed/worker-1
make -j$(nproc) depend/cli && go run ./cmd/atmosfeed-server worker --working-directory ~/.local/share/atmosfeed/var/lib/atmosfeed/worker-2
```

```shell
# Feed Deployment
git clone https://github.com/pojntfx/bluesky-feeds.git ../bluesky-feeds

cd ../bluesky-feeds
make -j$(nproc) build/function
cd ../atmosfeed

export ATMOSFEED_PASSWORD='asdf'
export ATMOSFEED_USERNAME='pojntfxtesting.bsky.social'
export ATMOSFEED_ATMOSFEED_URL='http://localhost:1337'
export ATMOSFEED_FEED_GENERATOR_DID='did:web:atmosfeed.serveo.net'

go run ./cmd/atmosfeed-client/ apply --feed-rkey everything --feed-classifier ../bluesky-feeds/out/local-everything-latest.scale
go run ./cmd/atmosfeed-client/ publish --feed-rkey everything --feed-name 'Atmosfeed Everything' --feed-description 'Newest posts on Bluesky (testing feed)'

go run ./cmd/atmosfeed-client/ apply --feed-rkey questions --feed-classifier ../bluesky-feeds/out/local-questions-latest.scale
go run ./cmd/atmosfeed-client/ publish --feed-rkey questions --feed-name 'Atmosfeed Questions' --feed-description 'Most popular questions on Bluesky in the last 24h (testing feed).'

go run ./cmd/atmosfeed-client/ apply --feed-rkey german --feed-classifier ../bluesky-feeds/out/local-german-latest.scale
go run ./cmd/atmosfeed-client/ publish --feed-rkey german --feed-name 'Atmosfeed German' --feed-description 'Most popular German posts on Bluesky in the last 24h (testing feed)'

go run ./cmd/atmosfeed-client/ apply --feed-rkey trending --feed-classifier ../bluesky-feeds/out/local-trending-latest.scale
go run ./cmd/atmosfeed-client/ publish --feed-rkey trending --feed-name 'Atmosfeed Trending' --feed-description 'Most popular trending posts on Bluesky in the last 24h (testing feed)'
```

```shell
# Feed Cleanup
go run ./cmd/atmosfeed-client/ delete --feed-rkey questions
go run ./cmd/atmosfeed-client/ unpublish --feed-rkey questions

go run ./cmd/atmosfeed-client/ delete --feed-rkey german
go run ./cmd/atmosfeed-client/ unpublish --feed-rkey german

go run ./cmd/atmosfeed-client/ delete --feed-rkey everything
go run ./cmd/atmosfeed-client/ unpublish --feed-rkey everything

go run ./cmd/atmosfeed-client/ delete --feed-rkey trending
go run ./cmd/atmosfeed-client/ unpublish --feed-rkey trending
```

```shell
# Frontend
cd frontend
bun dev # Now visit http://localhost:3000 to open the frontend and sign in
```

```shell
# Privacy & Interoperability
go run ./cmd/atmosfeed-client/ export-userdata --out ../bluesky-feeds/out/atmosfeed-userdata
go run ./cmd/atmosfeed-client/ delete-userdata --username pojntfxtesting.bsky.social --password ${PASSWORD}
```

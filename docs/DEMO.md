# Atmosfeed Demo

```shell
# Dependencies
docker rm -f atmosfeed-postgres && docker run -d --name atmosfeed-postgres -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=atmosfeed postgres
docker rm -f atmosfeed-redis && docker run --name atmosfeed-redis -p 6379:6379 -d redis
docker rm -f atmosfeed-minio && docker run --name atmosfeed-minio -p 9000:9000 -d minio/minio server /data

make -j$(nproc) depend

export ATMOSFEED_ORIGIN='http://localhost:3000'
make -j$(nproc) depend/cli && go run ./cmd/atmosfeed-server manager

make -j$(nproc) depend/cli && go run ./cmd/atmosfeed-server worker --working-directory ~/.local/share/atmosfeed/var/lib/atmosfeed/worker-1
make -j$(nproc) depend/cli && go run ./cmd/atmosfeed-server worker --working-directory ~/.local/share/atmosfeed/var/lib/atmosfeed/worker-2

ssh -R atmosfeed-feeds.serveo.net:80:localhost:1337 serveo.net

# End-to-End deployment
make -j$(nproc) depend/classifier/questions

go run ./cmd/atmosfeed-client/ list --username felicitas.pojtinger.com --password=${PASSWORD}
go run ./cmd/atmosfeed-client/ apply --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions --feed-classifier out/local-questions-latest.scale
go run ./cmd/atmosfeed-client/ publish --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions --feed-name 'Atmosfeed Questions' --feed-description 'Most popular questions on Bluesky in the last 24h (testing feed).' --feed-generator-did 'did:web:atmosfeed-feeds.serveo.net'
go run ./cmd/atmosfeed-client/ list --username felicitas.pojtinger.com --password=${PASSWORD}
go run ./cmd/atmosfeed-client/ unpublish --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions
go run ./cmd/atmosfeed-client/ delete --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions
go run ./cmd/atmosfeed-client/ list --username felicitas.pojtinger.com --password=${PASSWORD}

# Building the classifiers
make -j$(nproc) depend/classifier

go run ./cmd/atmosfeed-client/ apply --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey everything --feed-classifier out/local-everything-latest.scale
go run ./cmd/atmosfeed-client/ publish --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey everything --feed-name 'Atmosfeed Everything' --feed-description 'Newest posts on Bluesky (testing feed)' --feed-generator-did 'did:web:atmosfeed-feeds.serveo.net'

go run ./cmd/atmosfeed-client/ apply --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions --feed-classifier out/local-questions-latest.scale
go run ./cmd/atmosfeed-client/ publish --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions --feed-name 'Atmosfeed Questions' --feed-description 'Most popular questions on Bluesky in the last 24h (testing feed).' --feed-generator-did 'did:web:atmosfeed-feeds.serveo.net'

go run ./cmd/atmosfeed-client/ apply --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey german --feed-classifier out/local-german-latest.scale
go run ./cmd/atmosfeed-client/ publish --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey german --feed-name 'Atmosfeed German' --feed-description 'Most popular German posts on Bluesky in the last 24h (testing feed)' --feed-generator-did 'did:web:atmosfeed-feeds.serveo.net'

go run ./cmd/atmosfeed-client/ apply --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey trending --feed-classifier out/local-trending-latest.scale
go run ./cmd/atmosfeed-client/ publish --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey trending --feed-name 'Atmosfeed Trending' --feed-description 'Most popular trending posts on Bluesky in the last 24h (testing feed)' --feed-generator-did 'did:web:atmosfeed-feeds.serveo.net'

# Cleanup for everything but trending
go run ./cmd/atmosfeed-client/ delete --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions
go run ./cmd/atmosfeed-client/ delete --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey german
go run ./cmd/atmosfeed-client/ delete --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey everything

cd frontend
bun dev # Now visit http://localhost:3000 to open the frontend and sign in
```

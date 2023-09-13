# Atmosfeed Demo

```shell
docker rm -f atmosfeed-postgres && docker run -d --name atmosfeed-postgres -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=atmosfeed postgres
docker rm -f atmosfeed-redis && docker run --name atmosfeed-redis -p 6379:6379 -d redis
docker rm -f atmosfeed-minio && docker run --name atmosfeed-minio -p 9000:9000 -d minio/minio server /data

make -j$(nproc) depend/sql && go run ./cmd/atmosfeed-manager

make -j$(nproc) depend/sql && go run ./cmd/atmosfeed-worker --working-directory ~/.local/share/atmosfeed/var/lib/atmosfeed/worker-1
make -j$(nproc) depend/sql && go run ./cmd/atmosfeed-worker --working-directory ~/.local/share/atmosfeed/var/lib/atmosfeed/worker-2

make -j$(nproc) depend/classifier/questions

go run ./cmd/atmosfeed-client/ list --username felicitas.pojtinger.com --password=${PASSWORD}
go run ./cmd/atmosfeed-client/ apply --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions --feed-classifier out/local-questions-latest.scale
go run ./cmd/atmosfeed-client/ publish --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions --feed-name 'Atmosfeed Questions' --feed-description 'Most popular questions on Bluesky in the last 24h (testing feed).' --feed-generator-did 'did:web:atmosfeed-feeds.serveo.net'
go run ./cmd/atmosfeed-client/ list --username felicitas.pojtinger.com --password=${PASSWORD}
go run ./cmd/atmosfeed-client/ --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions --delete
go run ./cmd/atmosfeed-client/ --username felicitas.pojtinger.com --password=${PASSWORD} --list

go run ./cmd/atmosfeed-client/ --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey questions --delete
go run ./cmd/atmosfeed-client/ --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey german --delete
go run ./cmd/atmosfeed-client/ --username felicitas.pojtinger.com --password=${PASSWORD} --feed-rkey everything --delete
```

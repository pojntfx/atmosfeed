#!/bin/bash

rm -rf /tmp/feed-generator
git clone https://github.com/bluesky-social/feed-generator.git /tmp/feed-generator
cp ./hack/publish-feed.ts /tmp/feed-generator
cd /tmp/feed-generator
npm install

export FEEDGEN_HOSTNAME='atmosfeed-feeds.serveo.net'
export HANDLE='felicitas.pojtinger.com'
# export PASSWORD=''
export RECORD_NAME='trending'
export DISPLAY_NAME='Atmosfeed Trending'

npx ts-node publish-feed.ts

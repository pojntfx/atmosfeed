import { AtUri, BskyAgent } from "@atproto/api";
import { IFeed } from "./models";

export class RestAPI {
  constructor(
    private apiURL: URL,
    private service: string,
    private accessJWT: string,
    private agent: BskyAgent,
    private did: string
  ) {}

  async getFeeds(): Promise<IFeed[]> {
    const atmosfeedURL = new URL(this.apiURL + "admin/feeds");

    atmosfeedURL.search = new URLSearchParams({
      service: this.service,
    }).toString();

    const atmosfeedFeeds = (await (
      await fetch(atmosfeedURL.toString(), {
        headers: {
          Authorization: "Bearer " + this.accessJWT,
        },
      })
    ).json()) as string[];

    const bskyFeeds = await this.agent.app.bsky.feed.getActorFeeds({
      actor: this.did,
    });

    if (!bskyFeeds.success) {
      throw new Error("could not fetch feeds from Bluesky");
    }

    return atmosfeedFeeds
      .map((v) => {
        const bskyFeed = bskyFeeds.data.feeds.find(
          (f) => new AtUri(f.uri).rkey === v
        );

        if (bskyFeed) {
          return {
            rkey: v,
            title: bskyFeed.displayName,
            description: bskyFeed.description,
          };
        }

        return {
          rkey: v,
        };
      })
      .sort((a, b) => (a.title || "").localeCompare(b.title || ""));
  }
}

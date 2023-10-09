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

  async getFeeds(): Promise<{ published: IFeed[]; unpublished: IFeed[] }> {
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

    return atmosfeedFeeds.reduce(
      (acc, v) => {
        const bskyFeed = bskyFeeds.data.feeds.find(
          (f) => new AtUri(f.uri).rkey === v
        );

        if (bskyFeed) {
          acc.published.push({
            rkey: v,
            title: bskyFeed.displayName,
            description: bskyFeed.description,
          });
        } else {
          acc.unpublished.push({
            rkey: v,
          });
        }

        return acc;
      },
      { published: [] as IFeed[], unpublished: [] as IFeed[] }
    );
  }

  async applyFeed(rkey: string, classifier: File) {
    const atmosfeedURL = new URL(this.apiURL + "admin/feeds");

    atmosfeedURL.search = new URLSearchParams({
      rkey,
      service: this.service,
    }).toString();

    await fetch(atmosfeedURL.toString(), {
      method: "PUT",
      body: classifier,
      headers: {
        Authorization: "Bearer " + this.accessJWT,
        "Content-Type": "application/octet-stream",
      },
    });
  }
}

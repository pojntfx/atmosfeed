import { AtUri, BskyAgent } from "@atproto/api";
import { IFeed, IStructuredUserdata } from "./models";

const lexiconFeedGenerator = "app.bsky.feed.generator";

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
      throw new Error(
        `could not fetch feeds from Bluesky: ${JSON.stringify(bskyFeeds)}`
      );
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

  async applyFeed(
    rkey: string,
    classifier: File,
    pinnedDID: string,
    pinnedRkey: string
  ) {
    const atmosfeedURL = new URL(this.apiURL + "admin/feeds");

    atmosfeedURL.search = new URLSearchParams({
      rkey,
      service: this.service,
      pinnedDID,
      pinnedRkey,
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

  async finalizeFeed(
    feedGeneratorDID: string,
    rkey: string,
    name: string,
    description: string
  ) {
    const res = await this.agent.com.atproto.repo.createRecord({
      collection: lexiconFeedGenerator,
      repo: this.did,
      rkey: rkey,

      record: {
        createdAt: new Date().toISOString(),
        description: description,
        did: feedGeneratorDID,
        displayName: name,
      },
    });

    if (!res.success) {
      throw new Error(
        `could not finalize feed on Bluesky: ${JSON.stringify(res)}`
      );
    }
  }

  async republishFeed(
    feedGeneratorDID: string,
    rkey: string,
    name: string,
    description: string
  ) {
    const oldFeed = await this.agent.com.atproto.repo.getRecord({
      collection: lexiconFeedGenerator,
      repo: this.did,
      rkey: rkey,
    });

    if (!oldFeed.success) {
      throw new Error(
        `could not fetch existing feed from Bluesky: ${JSON.stringify(oldFeed)}`
      );
    }

    const res = await this.agent.com.atproto.repo.putRecord({
      collection: lexiconFeedGenerator,
      repo: this.did,
      rkey: rkey,

      record: {
        createdAt: new Date().toISOString(),
        description: description,
        did: feedGeneratorDID,
        displayName: name,
      },

      swapRecord: oldFeed.data.cid,
    });

    if (!res.success) {
      throw new Error(
        `could not republish feed on Bluesky: ${JSON.stringify(res)}`
      );
    }
  }

  async deleteFeed(rkey: string) {
    const atmosfeedURL = new URL(this.apiURL + "admin/feeds");

    atmosfeedURL.search = new URLSearchParams({
      rkey,
      service: this.service,
    }).toString();

    await fetch(atmosfeedURL.toString(), {
      method: "DELETE",
      headers: {
        Authorization: "Bearer " + this.accessJWT,
      },
    });
  }

  async unpublishFeed(rkey: string) {
    const res = await this.agent.com.atproto.repo.deleteRecord({
      collection: lexiconFeedGenerator,
      repo: this.did,
      rkey: rkey,
    });

    if (!res.success) {
      throw new Error(
        `could not unpublish feed from Bluesky: ${JSON.stringify(res)}`
      );
    }
  }

  async deleteUserdata() {
    const atmosfeedURL = new URL(this.apiURL + "userdata");

    atmosfeedURL.search = new URLSearchParams({
      service: this.service,
    }).toString();

    await fetch(atmosfeedURL.toString(), {
      method: "DELETE",
      headers: {
        Authorization: "Bearer " + this.accessJWT,
      },
    });
  }

  async exportStructuredUserdata(): Promise<IStructuredUserdata> {
    const atmosfeedURL = new URL(this.apiURL + "userdata/structured");

    atmosfeedURL.search = new URLSearchParams({
      service: this.service,
    }).toString();

    return (await (
      await fetch(atmosfeedURL.toString(), {
        headers: {
          Authorization: "Bearer " + this.accessJWT,
        },
      })
    ).json()) as IStructuredUserdata;
  }

  async exportClassifier(rkey: string): Promise<Blob> {
    const atmosfeedURL = new URL(this.apiURL + "userdata/blob");

    atmosfeedURL.search = new URLSearchParams({
      service: this.service,
      resource: "classifier",
      rkey,
    }).toString();

    return (
      await fetch(atmosfeedURL.toString(), {
        headers: {
          Authorization: "Bearer " + this.accessJWT,
        },
      })
    ).blob();
  }
}

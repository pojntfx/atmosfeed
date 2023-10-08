import { IFeed } from "./models";

export class RestAPI {
  constructor(
    private apiURL: URL,
    private service: string,
    private accessJWT: string
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

    return atmosfeedFeeds.map((v) => ({
      rkey: v,
      published: false,
    }));
  }
}

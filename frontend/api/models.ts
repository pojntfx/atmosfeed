export interface IFeed {
  rkey: string;
  title?: string;
  description?: string;
}

export interface IStructuredUserdata {
  feeds?: IStructuredUserdataFeed[];
  posts?: IStructuredUserdataPost[];
  feedPosts?: IStructuredUserdataFeedPost[];
}

export interface IStructuredUserdataFeed {
  Did: string;
  Rkey: string;
}

export interface IStructuredUserdataPost {
  Did: string;
  Rkey: string;
  CreatedAt: string;
  Text: string;
  Reply: boolean;
  Langs: string[];
  Likes: number;
}

export interface IStructuredUserdataFeedPost {
  FeedDid: string;
  FeedRkey: string;
  PostDid: string;
  PostRkey: string;
  Weight: number;
}

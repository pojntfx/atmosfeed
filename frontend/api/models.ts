export interface IFeedMetadata {
  rkey: string;
  pinnedDID: string;
  pinnedRkey: string;
}

export interface IFeed {
  rkey: string;
  title?: string;
  description?: string;
  pinnedPost?: string;
}

export interface IStructuredUserdata {
  feeds?: IStructuredUserdataFeed[];
  posts?: IStructuredUserdataPost[];
  feedPosts?: IStructuredUserdataFeedPost[];
}

export interface IStructuredUserdataFeed {
  did: string;
  rkey: string;
}

export interface IStructuredUserdataPost {
  did: string;
  rkey: string;
  createdAt: string;
  text: string;
  reply: boolean;
  langs: string[];
  likes: number;
}

export interface IStructuredUserdataFeedPost {
  feedDid: string;
  feedRkey: string;
  postDID: string;
  postRkey: string;
  weight: number;
}

import { IFeed } from "@/api/models";
import { RestAPI } from "@/api/rest";
import { BskyAgent } from "@atproto/api";
import { useCallback, useState } from "react";
import { useAsyncEffect } from "use-async-effect";

export const useAPI = (
  username: string,
  appPassword: string,

  service: string,
  atmosfeedAPI: string,

  clearAppPassword: () => void,

  handleError: (err: Error, loggedOut: boolean) => void
) => {
  const [agent, setAgent] = useState<BskyAgent>();
  const [avatar, setAvatar] = useState("");
  const [loading, setLoading] = useState(true);
  const [did, setDID] = useState("");
  const [accessJWT, setAccessJWT] = useState("");

  const logout = useCallback(() => {
    setAPI(undefined);
    clearAppPassword();
  }, [clearAppPassword]);

  useAsyncEffect(async () => {
    if (!username || !appPassword || !service) {
      setAvatar("");

      setLoading(false);

      return;
    }

    setLoading(true);

    const agent = new BskyAgent({
      service,
    });

    try {
      const res = await agent.login({
        identifier: username,
        password: appPassword,
      });

      setDID(res.data.did);
      setAccessJWT(res.data.accessJwt);
    } catch (e) {
      handleError(e as Error, true);

      logout();
    }

    setAgent(agent);
  }, [username, appPassword, service]);

  useAsyncEffect(async () => {
    if (!agent) {
      setAvatar("");

      return;
    }

    try {
      setAvatar(
        (
          await agent.getProfile({
            actor: username,
          })
        ).data.avatar || ""
      );
    } catch (e) {
      handleError(e as Error, true);

      logout();
    }
  }, [agent]);

  const [api, setAPI] = useState<RestAPI>();
  useAsyncEffect(() => {
    if (!atmosfeedAPI || !service || !accessJWT || !agent || !did) {
      return;
    }

    setAPI(new RestAPI(new URL(atmosfeedAPI), service, accessJWT, agent, did));
  }, [atmosfeedAPI, service, accessJWT, agent, did]);

  const [unpublishedFeeds, setUnpublishedFeeds] = useState<IFeed[]>([]);
  const [publishedFeeds, setPublishedFeeds] = useState<IFeed[]>([]);
  useAsyncEffect(async () => {
    if (!api) {
      return;
    }

    setLoading(true);

    try {
      const res = await api.getFeeds();

      setUnpublishedFeeds(res.unpublished);
      setPublishedFeeds(res.published);
    } catch (e) {
      handleError(e as Error, false);
    } finally {
      setLoading(false);
    }
  }, [api]);

  return {
    avatar,
    did,
    signedIn: api ? true : false,

    unpublishedFeeds,
    publishedFeeds,

    applyFeed: async (
      rkey: string,
      classifier: File,
      pinnedDID: string,
      pinnedRkey: string
    ) => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        await api.applyFeed(rkey, classifier, pinnedDID, pinnedRkey);

        const res = await api.getFeeds();

        setUnpublishedFeeds(res.unpublished);
        setPublishedFeeds(res.published);
      } catch (e) {
        handleError(e as Error, false);

        throw e;
      } finally {
        setLoading(false);
      }
    },
    patchFeed: async (rkey: string, pinnedDID: string, pinnedRkey: string) => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        await api.patchFeed(rkey, pinnedDID, pinnedRkey);

        const res = await api.getFeeds();

        setUnpublishedFeeds(res.unpublished);
        setPublishedFeeds(res.published);
      } catch (e) {
        handleError(e as Error, false);

        throw e;
      } finally {
        setLoading(false);
      }
    },
    finalizeFeed: async (
      feedGeneratorDID: string,
      rkey: string,
      name: string,
      description: string
    ) => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        await api.finalizeFeed(feedGeneratorDID, rkey, name, description);

        const res = await api.getFeeds();

        setUnpublishedFeeds(res.unpublished);
        setPublishedFeeds(res.published);
      } catch (e) {
        handleError(e as Error, false);

        throw e;
      } finally {
        setLoading(false);
      }
    },
    republishFeed: async (
      feedGeneratorDID: string,
      rkey: string,
      name: string,
      description: string
    ) => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        await api.republishFeed(feedGeneratorDID, rkey, name, description);

        const res = await api.getFeeds();

        setUnpublishedFeeds(res.unpublished);
        setPublishedFeeds(res.published);
      } catch (e) {
        handleError(e as Error, false);

        throw e;
      } finally {
        setLoading(false);
      }
    },
    deleteClassifier: async (rkey: string) => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        await api.deleteFeed(rkey);

        const res = await api.getFeeds();

        setUnpublishedFeeds(res.unpublished);
        setPublishedFeeds(res.published);
      } catch (e) {
        handleError(e as Error, false);

        throw e;
      } finally {
        setLoading(false);
      }
    },
    unpublishFeed: async (rkey: string) => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        await api.unpublishFeed(rkey);

        const res = await api.getFeeds();

        setUnpublishedFeeds(res.unpublished);
        setPublishedFeeds(res.published);
      } catch (e) {
        handleError(e as Error, false);

        throw e;
      } finally {
        setLoading(false);
      }
    },
    deleteFeed: async (rkey: string) => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        await api.unpublishFeed(rkey);
        await api.deleteFeed(rkey);

        const res = await api.getFeeds();

        setUnpublishedFeeds(res.unpublished);
        setPublishedFeeds(res.published);
      } catch (e) {
        handleError(e as Error, false);

        throw e;
      } finally {
        setLoading(false);
      }
    },

    deleteUserdata: async () => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        await api.deleteUserdata();

        logout();
      } catch (e) {
        handleError(e as Error, false);

        throw e;
      } finally {
        setLoading(false);
      }
    },
    exportUserdata: async () => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        const structuredUserdata = await api.exportStructuredUserdata();

        const data = {
          did,
          service,

          structured: structuredUserdata,
        };

        const blob = new Blob([JSON.stringify(data, null, 2)], {
          type: "application/json",
        });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");

        a.href = url;
        a.download = "atmosfeed.json";
        a.click();

        URL.revokeObjectURL(url);

        if (structuredUserdata.feeds) {
          await Promise.all(
            structuredUserdata.feeds.map(async (f) => {
              const classifier = await api.exportClassifier(f.Rkey);

              const url = URL.createObjectURL(classifier);
              const a = document.createElement("a");

              a.href = url;
              a.download = f.Rkey + ".scale";
              a.click();

              URL.revokeObjectURL(url);
            })
          );
        }
      } catch (e) {
        handleError(e as Error, false);

        throw e;
      } finally {
        setLoading(false);
      }
    },

    loading,
    logout,
  };
};

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

    deleteData: async () => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        // TODO: Call API to delete all user data

        logout();
      } catch (e) {
        handleError(e as Error, false);
      } finally {
        setLoading(false);
      }
    },

    loading,
    logout,
  };
};

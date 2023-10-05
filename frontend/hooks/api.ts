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
  const [refreshJWT, setRefreshJWT] = useState("");

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
      setRefreshJWT(res.data.refreshJwt);
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

  const [api, setAPI] = useState<any>();
  useAsyncEffect(() => {
    if (!atmosfeedAPI || !service || !accessJWT || !refreshJWT) {
      return;
    }

    setAPI(true);
  }, [atmosfeedAPI, service, accessJWT, refreshJWT]);

  useAsyncEffect(async () => {
    if (!api) {
      return;
    }

    setLoading(true);

    try {
      // const res = await api.getConfiguration();
      // setPostTTL(res.postTTL);
      // setEnabled(res.enabled);
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

    deleteData: async () => {
      if (!api) {
        return;
      }

      setLoading(true);

      try {
        // await api.deleteConfiguration();

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

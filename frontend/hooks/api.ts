import { BskyAgent } from "@atproto/api";
import { useState } from "react";
import { useAsyncEffect } from "use-async-effect";

export const useAPI = (
  username: string,
  appPassword: string,

  service: string,
  atmosfeedAPI: string,

  logout: () => void
) => {
  const [agent, setAgent] = useState<BskyAgent>();

  useAsyncEffect(async () => {
    if (!username || !appPassword || !service) {
      return;
    }

    const agent = new BskyAgent({
      service,
    });

    try {
      await agent.login({
        identifier: username,
        password: appPassword,
      });
    } catch (e) {
      console.error(e);

      logout();
    }

    setAgent(agent);
  }, [username, appPassword, service]);

  const [avatar, setAvatar] = useState("");
  useAsyncEffect(async () => {
    if (!agent) {
      return;
    }

    setAvatar(
      (
        await agent.getProfile({
          actor: username,
        })
      ).data.avatar || ""
    );
  }, [agent]);

  return {
    avatar,
  };
};

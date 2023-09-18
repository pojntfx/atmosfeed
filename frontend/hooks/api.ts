import { BskyAgent } from "@atproto/api";
import { useState } from "react";
import { useAsyncEffect } from "use-async-effect";

export const useAPI = (
  service: string,
  username: string,
  appPassword: string,

  logout: () => void
) => {
  const [agent, setAgent] = useState<BskyAgent>();

  useAsyncEffect(async () => {
    if (!service || !username || !appPassword) {
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
  }, [service, username, appPassword]);

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

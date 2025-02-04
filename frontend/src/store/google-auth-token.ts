import { create } from "zustand";
import { repository } from "~wails/models";
import {
  GetGoogleAuthToken,
  DeleteGoogleAuthToken,
  RefreshGoogleAuthToken,
} from "~wails/main/App";

type GoogleAuthTokenState = {
  token: repository.GoogleAuthToken | null;
  getToken: () => Promise<repository.GoogleAuthToken | null>;
  refreshToken: () => Promise<void>;
  deleteToken: () => Promise<void>;
};

export function isGoogleAPIInvalidGrantError(error: any) {
  let message = typeof error === "string" ? error : error.message;
  return message.includes("invalid_grant") && message.includes("oauth2");
}

export const useGoogleAuthTokenStore = create<GoogleAuthTokenState>((set) => ({
  token: null,

  refreshToken: async () => {
    const token = await RefreshGoogleAuthToken();
    set({ token });
  },

  getToken: async () => {
    const token = await GetGoogleAuthToken();
    set({ token });

    return token;
  },

  deleteToken: async () => {
    await DeleteGoogleAuthToken();
    set({ token: null });
  },
}));

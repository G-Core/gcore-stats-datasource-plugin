/**
 * Helpers for working with API tokens and Authorization headers.
 */
export enum AuthSchema {
  APIKey = "APIKey",
  Bearer = "Bearer",
}

export const isJWT = (token: string): boolean => {
  try {
    return JSON.parse(atob(token.split(".")[0])).typ === "JWT";
  } catch (e) {
    return false;
  }
};

export const haveAuthSchema = (target: string): boolean => {
  try {
    if (!target || !target.trim()) {
      return false;
    }
    const schema = target.trim().split(" ");
    return schema[0] in AuthSchema;
  } catch (e) {
    return false;
  }
};

export const getAuthorizationValue = (token: string): string => {
  if (token.length === 0) {
    return "";
  }

  if (!haveAuthSchema(token)) {
    return isJWT(token)
      ? `${AuthSchema.Bearer} ${token}`
      : `${AuthSchema.APIKey} ${token}`;
  }
  return token;
};

export const stripTrailingSlash = (url: string): string => {
  return url.endsWith("/") ? url.slice(0, -1) : url;
};

export const removeHttpPrefix = (url: string): string =>
  url.replace(/^https?:\/\//, "");

/** Hostname segment only: trims, strips http(s):// and one trailing slash. */
export const normalizeApiUrlInput = (raw: string): string => {
  return stripTrailingSlash(removeHttpPrefix(raw.trim()));
};

const hostnameWithoutPort = (host: string): string => {
  if (!host.includes(":")) {
    return host;
  }
  const lastColon = host.lastIndexOf(":");
  const tail = host.slice(lastColon + 1);
  if (/^\d+$/.test(tail)) {
    return host.slice(0, lastColon);
  }
  return host;
};

/**
 * Empty string is valid (backend uses default api host).
 * Otherwise: no "/", hostname (ignoring port) must start with api. and end with .com (case-insensitive).
 */
export const isValidApiHostname = (normalized: string): boolean => {
  if (normalized === "") {
    return true;
  }
  if (normalized.includes("/")) {
    return false;
  }
  const host = hostnameWithoutPort(normalized).toLowerCase();
  return host.startsWith("api.") && host.endsWith(".com");
};

export const getHostnameValue = (url: string): string => {
  return normalizeApiUrlInput(url);
};

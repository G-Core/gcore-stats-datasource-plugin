import {
  AuthSchema,
  getAuthorizationValue,
  getHostnameValue,
  haveAuthSchema,
  isJWT,
  isValidApiHostname,
  normalizeApiUrlInput,
} from "token";

const validJWTToken =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c";

describe("Token", () => {
  it("should return false if not token is invalid", () => {
    expect(isJWT("")).toBe(false);
    expect(isJWT("00000000.00000.0000")).toBe(false);
    expect(isJWT("sadasdadasd.asdadadasd.asdasdd")).toBe(false);
    expect(isJWT("asdadadadadda")).toBe(false);
  });

  it("should return true if not token is valid", () => {
    expect(isJWT(validJWTToken)).toBe(true);
  });

  it("should return true for Bearer token", () => {
    expect(haveAuthSchema(`Bearer ${validJWTToken}`)).toBe(true);
    expect(haveAuthSchema(`APIKey ${validJWTToken}`)).toBe(true);
  });

  it("should return true for valid Bearer token or APIKey token", () => {
    expect(haveAuthSchema(`${AuthSchema.Bearer} ${validJWTToken}`)).toBe(true);
    expect(haveAuthSchema(`${AuthSchema.APIKey} ${validJWTToken}`)).toBe(true);
  });

  it("should return false if not include valid auth token ", () => {
    expect(haveAuthSchema("   ")).toBe(false);
    expect(haveAuthSchema(`${validJWTToken}`)).toBe(false);
    expect(haveAuthSchema("asdasddasd")).toBe(false);
    expect(haveAuthSchema(`apikey ${validJWTToken}`)).toBe(false);
    expect(haveAuthSchema(`ApiKey ${validJWTToken}`)).toBe(false);
    expect(haveAuthSchema(`${validJWTToken} ${validJWTToken}`)).toBe(false);
  });

  it("should return value Authorization value", () => {
    expect(getAuthorizationValue("")).toBe("");
    expect(getAuthorizationValue(validJWTToken)).toBe(
      `${AuthSchema.Bearer} ${validJWTToken}`
    );
    expect(getAuthorizationValue("sdfsdf")).toBe(`${AuthSchema.APIKey} sdfsdf`);
    expect(getAuthorizationValue(`${AuthSchema.Bearer} ${validJWTToken}`)).toBe(
      `${AuthSchema.Bearer} ${validJWTToken}`
    );
    expect(getAuthorizationValue(`${AuthSchema.APIKey} ${validJWTToken}`)).toBe(
      `${AuthSchema.APIKey} ${validJWTToken}`
    );
  });

  it("should accept valid API hostnames only", () => {
    expect(isValidApiHostname("")).toBe(true);
    expect(isValidApiHostname("api.gcore.com")).toBe(true);
    expect(isValidApiHostname("API.GCORE.COM")).toBe(true);
    expect(isValidApiHostname("api.staging.gcore.com")).toBe(true);
    expect(isValidApiHostname("api.gcore.com:443")).toBe(true);
    expect(normalizeApiUrlInput("https://api.gcore.com/")).toBe("api.gcore.com");
    expect(getHostnameValue("https://api.gcore.com/")).toBe("api.gcore.com");
    expect(isValidApiHostname(getHostnameValue("https://api.gcore.com/"))).toBe(
      true
    );
  });

  it("should reject invalid API hostnames", () => {
    expect(isValidApiHostname("cdn.gcore.com")).toBe(false);
    expect(isValidApiHostname("api.gcore.com/foo")).toBe(false);
    expect(isValidApiHostname("https://evil.com")).toBe(false);
    expect(isValidApiHostname("evil.com")).toBe(false);
    expect(isValidApiHostname("api.gcore.org")).toBe(false);
    expect(isValidApiHostname("notapi.gcore.com")).toBe(false);
  });
});

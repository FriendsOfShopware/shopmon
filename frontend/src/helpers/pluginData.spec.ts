import { describe, expect, it } from "vitest";
import { parsePluginData } from "./pluginData";

describe("parsePluginData", () => {
  const validObj = {
    url: "https://shop.example.com",
    clientId: "SWIA123456789",
    clientSecret: "SecretKey1234567890",
  };
  const validJson = JSON.stringify(validObj);
  const validBase64 = btoa(validJson);

  it("throws EMPTY_INPUT for empty string or empty quotes", () => {
    expect(() => parsePluginData("   ")).toThrow("EMPTY_INPUT");
    expect(() => parsePluginData('""')).toThrow("EMPTY_INPUT");
  });

  it("parses valid base64 string", () => {
    const res = parsePluginData(validBase64);
    expect(res).toEqual(validObj);
  });

  it("parses base64 with internal newlines and spaces", () => {
    const formatted = `${validBase64.slice(0, 10)}\n  ${validBase64.slice(10, 20)}\r\n${validBase64.slice(20)}`;
    const res = parsePluginData(formatted);
    expect(res).toEqual(validObj);
  });

  it("parses base64 with surrounding quotes", () => {
    const res = parsePluginData(`"${validBase64}"`);
    expect(res).toEqual(validObj);
  });

  it("parses user example string with quotes and escaped slashes in json", () => {
    const example = `"eyJ1cmwiOiJodHRwOlwvXC9zdy10cnVuay5sb2NhbGhvc3QiLCJjbGllbnRJZCI6IlNXSUFFV1M0VDFWR09WUEdDRk5OVkhCWlpHIiwiY2xpZW50U2VjcmV0IjoiV2pCQmNtTjVSakZQWjBNMGR6TktZbFZYUzBSUGJEWklVRXRaTjFJNFkwUmxSazlyWjNVIn0="`;
    const res = parsePluginData(example);
    expect(res).toEqual({
      url: "http://sw-trunk.localhost",
      clientId: "SWIAEWS4T1VGOVPGCFNNVHBZZG",
      clientSecret: "WjBBcmN5RjFPZ0M0dzNKYlVXS0RPbDZIUEtZN1I4Y0RlRk9rZ3U",
    });
  });

  it("parses URL-safe base64 string and handles missing padding", () => {
    // btoa output with characters replaced
    const base64Url = validBase64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
    const res = parsePluginData(base64Url);
    expect(res).toEqual(validObj);
  });

  it("parses direct JSON string", () => {
    const res = parsePluginData(validJson);
    expect(res).toEqual(validObj);
  });

  it("supports snake_case keys in JSON", () => {
    const snakeObj = {
      shop_url: "https://shop.example.com",
      client_id: "SWIA123456789",
      client_secret: "SecretKey1234567890",
    };
    const res = parsePluginData(JSON.stringify(snakeObj));
    expect(res).toEqual(validObj);
  });

  it("throws INVALID_DATA if required fields are missing", () => {
    const incomplete = btoa(JSON.stringify({ url: "https://example.com" }));
    expect(() => parsePluginData(incomplete)).toThrow("INVALID_DATA");
  });

  it("throws INVALID_FORMAT for malformed input", () => {
    expect(() => parsePluginData("!!!not-base64!!!")).toThrow("INVALID_FORMAT");
  });
});

import { describe, expect, it } from "vitest";
import {
  base64URLToBuffer,
  bufferToBase64URL,
  parseCreationOptions,
  webAuthnErrorMessage,
} from "@/lib/webauthn";

describe("webauthn helpers", () => {
  it("round-trips base64url buffers", () => {
    const original = new Uint8Array([1, 2, 3, 4]).buffer;
    expect(base64URLToBuffer(bufferToBase64URL(original))).toEqual(original);
  });

  it("unwraps publicKey creation options from the API", () => {
    const challenge = bufferToBase64URL(new Uint8Array([9, 9, 9]).buffer);
    const userId = bufferToBase64URL(new Uint8Array([1, 2, 3]).buffer);

    const parsed = parseCreationOptions({
      publicKey: {
        rp: { name: "SnackMates", id: "localhost" },
        user: { id: userId, name: "user@example.com", displayName: "User" },
        challenge,
        pubKeyCredParams: [{ alg: -7, type: "public-key" }],
      },
    });

    expect(new Uint8Array(parsed.challenge as ArrayBuffer)).toEqual(new Uint8Array([9, 9, 9]));
    expect(new Uint8Array(parsed.user.id)).toEqual(new Uint8Array([1, 2, 3]));
  });

  it("rejects malformed creation options", () => {
    expect(() => parseCreationOptions({ publicKey: {} as never })).toThrow(
      "Invalid WebAuthn registration options from server.",
    );
  });

  it("maps cancelled WebAuthn prompts to a clear message", () => {
    expect(webAuthnErrorMessage(new DOMException("cancelled", "NotAllowedError"))).toBe(
      "Security key registration was cancelled. No key was registered.",
    );
    expect(webAuthnErrorMessage(new DOMException("aborted", "AbortError"))).toBe(
      "Security key registration was cancelled. No key was registered.",
    );
  });
});

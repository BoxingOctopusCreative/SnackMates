type Base64Like = string | ArrayBuffer;

export type WebAuthnCreationOptionsPayload = {
  publicKey?: PublicKeyCredentialCreationOptions & {
    challenge: Base64Like;
    user: PublicKeyCredentialUserEntity & { id: Base64Like };
    excludeCredentials?: PublicKeyCredentialDescriptor[];
  };
} & Partial<
  PublicKeyCredentialCreationOptions & {
    challenge: Base64Like;
    user: PublicKeyCredentialUserEntity & { id: Base64Like };
  }
>;

export function parseCreationOptions(
  options: WebAuthnCreationOptionsPayload,
): PublicKeyCredentialCreationOptions {
  const source = options.publicKey ?? options;
  const { rp, pubKeyCredParams, user, challenge } = source;
  if (!challenge || !user?.id || !rp || !pubKeyCredParams?.length) {
    throw new Error("Invalid WebAuthn registration options from server.");
  }

  return {
    rp,
    pubKeyCredParams,
    challenge: toBuffer(challenge),
    user: {
      ...user,
      id: toBuffer(user.id),
    },
    attestation: source.attestation,
    authenticatorSelection: source.authenticatorSelection,
    excludeCredentials: source.excludeCredentials?.map((cred) => ({
      ...cred,
      id: toBuffer(cred.id as unknown as Base64Like),
    })),
    extensions: source.extensions,
    timeout: source.timeout,
  };
}

function toBuffer(value: Base64Like): ArrayBuffer {
  if (value instanceof ArrayBuffer) {
    return value;
  }
  if (typeof value !== "string" || value.length === 0) {
    throw new Error("Expected a base64url-encoded WebAuthn value.");
  }
  return base64URLToBuffer(value);
}

export function bufferToBase64URL(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  bytes.forEach((b) => (binary += String.fromCharCode(b)));
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

export function webAuthnErrorMessage(err: unknown): string {
  if (err instanceof DOMException) {
    if (err.name === "NotAllowedError" || err.name === "AbortError") {
      return "Security key registration was cancelled. No key was registered.";
    }
    if (err.message) {
      return err.message;
    }
  }
  if (err instanceof Error && err.message) {
    return err.message;
  }
  return "WebAuthn registration failed.";
}

export function base64URLToBuffer(value: string): ArrayBuffer {
  const padded = value.replace(/-/g, "+").replace(/_/g, "/") + "===".slice((value.length + 3) % 4);
  const binary = atob(padded);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes.buffer;
}

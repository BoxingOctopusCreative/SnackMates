"use client";

import Script from "next/script";
import { useCallback, useEffect, useRef } from "react";
import { turnstileSiteKey } from "@/lib/turnstile";

type TurnstileRenderOptions = {
  sitekey: string;
  callback: (token: string) => void;
  "expired-callback"?: () => void;
  "error-callback"?: () => void;
};

type TurnstileAPI = {
  render: (container: HTMLElement, options: TurnstileRenderOptions) => string;
  remove: (widgetId: string) => void;
};

declare global {
  interface Window {
    turnstile?: TurnstileAPI;
  }
}

const TURNSTILE_SCRIPT = "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";

type TurnstileWidgetProps = {
  onToken: (token: string) => void;
  onExpire?: () => void;
  onError?: () => void;
  resetKey?: number;
};

export function TurnstileWidget({
  onToken,
  onExpire,
  onError,
  resetKey = 0,
}: TurnstileWidgetProps) {
  const siteKey = turnstileSiteKey();
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetIdRef = useRef<string | null>(null);
  const scriptReadyRef = useRef(false);

  const renderWidget = useCallback(() => {
    if (!siteKey || !containerRef.current || !window.turnstile) {
      return;
    }
    if (widgetIdRef.current) {
      window.turnstile.remove(widgetIdRef.current);
      widgetIdRef.current = null;
    }
    widgetIdRef.current = window.turnstile.render(containerRef.current, {
      sitekey: siteKey,
      callback: onToken,
      "expired-callback": () => onExpire?.(),
      "error-callback": () => onError?.(),
    });
  }, [siteKey, onToken, onExpire, onError]);

  useEffect(() => {
    if (scriptReadyRef.current) {
      renderWidget();
    }
    return () => {
      if (widgetIdRef.current && window.turnstile) {
        window.turnstile.remove(widgetIdRef.current);
        widgetIdRef.current = null;
      }
    };
  }, [renderWidget, resetKey]);

  if (!siteKey) {
    return null;
  }

  return (
    <>
      <Script
        src={TURNSTILE_SCRIPT}
        strategy="afterInteractive"
        onReady={() => {
          scriptReadyRef.current = true;
          renderWidget();
        }}
      />
      <div ref={containerRef} />
    </>
  );
}

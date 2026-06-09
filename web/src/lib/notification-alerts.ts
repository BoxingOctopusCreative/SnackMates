import { ToastQueue } from "@react-spectrum/toast";
import type { NotificationItem } from "@/lib/notifications";

let audioContext: AudioContext | null = null;

export function playNotificationSound() {
  if (typeof window === "undefined") return;

  try {
    audioContext ??= new AudioContext();
    if (audioContext.state === "suspended") {
      void audioContext.resume();
    }

    const ctx = audioContext;
    const now = ctx.currentTime;
    const gain = ctx.createGain();
    gain.connect(ctx.destination);
    gain.gain.setValueAtTime(0.0001, now);
    gain.gain.exponentialRampToValueAtTime(0.12, now + 0.02);
    gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.45);

    const first = ctx.createOscillator();
    first.type = "sine";
    first.frequency.setValueAtTime(880, now);
    first.connect(gain);
    first.start(now);
    first.stop(now + 0.12);

    const second = ctx.createOscillator();
    second.type = "sine";
    second.frequency.setValueAtTime(1174.66, now + 0.1);
    second.connect(gain);
    second.start(now + 0.1);
    second.stop(now + 0.45);
  } catch {
    // Browsers may block audio until the user has interacted with the page.
  }
}

export function notifySnackMateAccepted(displayName: string, username?: string) {
  const message = `${displayName} accepted your snack mate request!`;
  ToastQueue.positive(message, {
    timeout: 6000,
    ...(username
      ? {
          actionLabel: "View profile",
          shouldCloseOnAction: true,
          onAction: () => {
            window.location.assign(`/users/${username}`);
          },
        }
      : {}),
  });
  playNotificationSound();
}

export function notificationKey(item: NotificationItem) {
  return `${item.type}:${item.id}`;
}

export function findNewAcceptedNotifications(
  previous: Set<string>,
  items: NotificationItem[],
  initialized: boolean,
) {
  if (!initialized) return [];

  return items.filter(
    (item) => item.type === "snack_mate_accepted" && !previous.has(notificationKey(item)),
  );
}

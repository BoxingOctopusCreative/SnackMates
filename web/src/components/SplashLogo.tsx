import Image from "next/image";

const LOGO_URL = "https://assets.snackmates.food/brand/logokit_wide.png";

export function SplashLogo() {
  return (
    <Image
      src={LOGO_URL}
      alt="SnackMates"
      width={660}
      height={156}
      priority
      style={{ width: "min(100%, 36rem)", height: "auto" }}
    />
  );
}

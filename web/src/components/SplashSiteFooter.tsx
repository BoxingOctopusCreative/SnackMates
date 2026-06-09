const BOXING_OCTOPUS_URL = "https://boxingoctop.us";

export function SplashSiteFooter() {
  const year = new Date().getFullYear();

  return (
    <footer className="sm-splash-site-footer">
      A{" "}
      <a href={BOXING_OCTOPUS_URL} target="_blank" rel="noopener noreferrer">
        Boxing Octopus Creative
      </a>{" "}
      project. Copyright {year} All rights reserved.
    </footer>
  );
}

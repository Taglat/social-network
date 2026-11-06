/** @type {import('next').NextConfig} */
const isExport = process.env.NEXT_OUTPUT_EXPORT === "true";

const nextConfig = {
  reactStrictMode: true,
  // Produce a standalone server bundle for Docker, or export static files for GitLab Pages
  output: isExport ? "export" : "standalone",
  ...(isExport
    ? {
        images: { unoptimized: true },
        trailingSlash: true,
        basePath: process.env.PAGES_BASE_PATH || undefined,
      }
    : {}),
};

module.exports = nextConfig;

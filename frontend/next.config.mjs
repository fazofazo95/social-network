/** @type {import('next').NextConfig} */
const nextConfig = {
  allowedDevOrigins: ["localhost", "127.0.0.1", "*.localhost", "devtools"],
  experimental: {
    turbopackFileSystemCacheForDev: true,
  },
  /* config options here */
  reactCompiler: true,
};

export default nextConfig;

/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    turbopackFileSystemCacheForDev: true,
  },
  /* config options here */
  reactCompiler: true,
};

export default nextConfig;

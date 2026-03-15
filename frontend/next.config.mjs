/** @type {import('next').NextConfig} */
const isDev = process.env.NODE_ENV !== "production";
const apiBaseUrl =
  process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

const remotePatterns = [];

try {
  const apiUrl = new URL(apiBaseUrl);
  if (apiUrl.protocol === "http:" || apiUrl.protocol === "https:") {
    remotePatterns.push({
      protocol: apiUrl.protocol.replace(":", ""),
      hostname: apiUrl.hostname,
      port: apiUrl.port,
      pathname: "/uploads/**",
    });
  }
} catch {}

if (isDev) {
  remotePatterns.push(
    {
      protocol: "http",
      hostname: "localhost",
      port: "8080",
      pathname: "/uploads/**",
    },
    {
      protocol: "http",
      hostname: "127.0.0.1",
      port: "8080",
      pathname: "/uploads/**",
    },
  );
}

const internalApiUrl = process.env.INTERNAL_API_URL || apiBaseUrl;

const nextConfig = {
  output: "standalone",
  allowedDevOrigins: ["localhost", "127.0.0.1", "*.localhost", "devtools"],
  images: {
    dangerouslyAllowLocalIP: true,
    remotePatterns,
  },
  async rewrites() {
    return [
      {
        source: "/uploads/:path*",
        destination: `${internalApiUrl}/uploads/:path*`,
      },
    ];
  },
  experimental: {
    turbopackFileSystemCacheForDev: true,
  },
  reactCompiler: true,
};

export default nextConfig;

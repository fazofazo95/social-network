/** @type {import('next').NextConfig} */
const isDev = process.env.NODE_ENV !== "production";
const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

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
} catch {
}

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
    }
  );
}

const nextConfig = {
  allowedDevOrigins: ["localhost", "127.0.0.1", "*.localhost", "devtools"],
  images: {
    dangerouslyAllowLocalIP: isDev,
    remotePatterns,
  },
  experimental: {
    turbopackFileSystemCacheForDev: true,
  },
  /* config options here */
  reactCompiler: true,
};

export default nextConfig;

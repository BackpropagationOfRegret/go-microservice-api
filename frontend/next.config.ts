import type { NextConfig } from "next";

// Server-side proxy: browser calls /api on the same origin → no CORS.
// API_URL is read at build time (Docker) or from .env.local (npm run dev).
const apiUrl =
  process.env.API_URL ??
  process.env.NEXT_PUBLIC_API_URL ??
  "http://localhost:8090";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${apiUrl}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;

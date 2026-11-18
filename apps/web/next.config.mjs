const API_URL = (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080").replace(/\/+$/, "");

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: "standalone",
  // Uploaded files are stored as host-independent "/files/..." URLs; the web
  // app proxies them to the API so images work on any host/port.
  async rewrites() {
    return [{ source: "/files/:path*", destination: `${API_URL}/files/:path*` }];
  },
};

export default nextConfig;

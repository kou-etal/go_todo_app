import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {//rewritesはpromise返す。
    return [
      {
        //componentsでfetch(/api/:path*)した場合http://localhost:18000/:path*にする。
        //TODO:ここはハードコードしない。
        source: "/api/:path*",
        destination: "http://localhost:18000/:path*",
      },
    ];
  },
};

export default nextConfig;

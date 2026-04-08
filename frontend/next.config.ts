import type { NextConfig } from "next";

//ローカルではlocalhost:18000、Vercelでは環境変数でURTを設定。
//これ便利記法
const apiUrl = process.env.API_URL ?? "http://localhost:18000";

const nextConfig: NextConfig = {
  async rewrites() {//rewritesはpromise返す。
    return [
      {
        //componentsでfetch(/api/:path*)した場合API_URLの値/:path*にする。
        source: "/api/:path*",
        destination: `${apiUrl}/:path*`,
      },
    ];
  },
};

export default nextConfig;

/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    optimizeCss: false
  },
  output: 'standalone',
  distDir: '.next',
  trailingSlash: false,
  reactStrictMode: true
}

module.exports = nextConfig 
/** @type {import('next').NextConfig} */
const nextConfig = {
	images: {
		domains: [
			"pub-ebaf17804d5146cd98dcfec2fae780af.r2.dev",
		],
		remotePatterns: [
			{
				protocol: "https",
				hostname: "pub-ebaf17804d5146cd98dcfec2fae780af.r2.dev",
			},
			{
				protocol: "https",
				hostname: "api.dicebear.com",
			},
			{
				protocol: "https",
				hostname: "**.onrender.com",
			},
			{
				protocol: "http",
				hostname: "localhost",
			},
		],
	},
	async headers() {
		return [
			{
				source: "/:path*",
				headers: [
					{ key: "X-Content-Type-Options", value: "nosniff" },
					{ key: "X-Frame-Options", value: "DENY" },
					{ key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
					{ key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
					{ key: "Strict-Transport-Security", value: "max-age=31536000; includeSubDomains" },
					{ key: "Content-Security-Policy", value: "frame-ancestors 'none'; base-uri 'self'; object-src 'none'" },
				],
			},
		];
	},
	async redirects() {
		return [
			{ source: '/dashboard/candidates', destination: '/candidates', permanent: false },
			{ source: '/dashboard/candidates/new', destination: '/candidates/new', permanent: false },
			{ source: '/dashboard/candidates/:id', destination: '/candidates/:id', permanent: false },
			{ source: '/dashboard/selections', destination: '/selections', permanent: false },
			{ source: '/dashboard/selections/:id', destination: '/selections/:id', permanent: false },
			{ source: '/dashboard/notifications', destination: '/notifications', permanent: false },
			{ source: '/dashboard/settings', destination: '/settings', permanent: false },
		];
	},
};

export default nextConfig;

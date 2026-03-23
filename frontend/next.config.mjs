/** @type {import('next').NextConfig} */
const nextConfig = {
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

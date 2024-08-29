import { Client } from '@neondatabase/serverless';
import { query } from './query';
import { validateSearchRequest } from './validation';

async function getClient(databaseUrl: string): Promise<Client> {
	if (databaseUrl === undefined) {
		throw new Error('DATABASE_URL is required');
	}

	const client = new Client(databaseUrl);
	await client.connect();
	return client;
}

async function fetchData(client: Client, queryParam: string, ctx: ExecutionContext): Promise<Response> {
	try {
		const results = await query(client, queryParam);
		ctx.waitUntil(client.end());
		return Response.json(results);
	} catch (error) {
		console.error('Error during fetch:', error);
		return new Response('An internal server error occurred', { status: 500 });
	}
}

async function handleSearchRequest(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
	const validation = validateSearchRequest(request);
	if (validation.error) {
		return new Response(validation.error.message, { status: validation.error.status });
	}

	const client = await getClient(env.DATABASE_URL);
	return await fetchData(client, validation.queryParam, ctx);
}

async function attemptR2Request(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
	const cache = caches.default;
	let response = await cache.match(request);
	if (response) {
		return response;
	}

	const url = new URL(request.url);
	const objectKey = url.pathname.slice(1);
	return await serveR2Object(request, env, objectKey);
}

async function serveR2Object(request: Request, env: Env, objectKey: string) {
	const object = await env.BUCKET.get(objectKey);
	if (!object) {
		return new Response('Not Found', { status: 404 });
	}

	let response = new Response(object.body, {
		headers: {
			'Content-Type': object.httpMetadata!.contentType || 'application/octet-stream',
			'Cache-Control': 'public, max-age=3600', // Cache for 1 hour
		},
	});
	const cache = caches.default;
	await cache.put(request, response.clone());
	return response;
}

export default {
	async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
		const url = new URL(request.url);

		switch (url.pathname) {
			case '/search':
				return await handleSearchRequest(request, env, ctx);
			case '/':
				return await serveR2Object(request, env, 'index.html');
			default:
				return await attemptR2Request(request, env, ctx);
		}
	},
};

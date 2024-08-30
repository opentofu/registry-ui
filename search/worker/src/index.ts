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

function applyCorsHeaders(response: Response) {
	response.headers.set('Access-Control-Allow-Origin', '*');
	response.headers.set('Access-Control-Allow-Methods', 'GET');
	return response;
}

async function serveR2Object(request: Request, env: Env, objectKey: string) {
	const cache = caches.default;
	let response = await cache.match(request);
	if (response) {
		return applyCorsHeaders(new Response(response.body, response));
	}

	const object = await env.BUCKET.get(objectKey);
	if (!object) {
		return new Response('Not Found', { status: 404 });
	}

	response = new Response(object.body, {
		headers: {
			'Content-Type': object.httpMetadata!.contentType || 'application/octet-stream',
			'Cache-Control': 'public, max-age=3600', // Cache for 1 hour
		},
	});
	await cache.put(request, response.clone());
	return applyCorsHeaders(response);
}

export default {
	async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
		// This is a readonly api right now, we only support GET requests
		if (request.method !== 'GET') {
			return new Response('Method Not Allowed', { status: 405 });
		}

		const url = new URL(request.url);

		let response: Response;
		switch (url.pathname) {
			case '/search':
				response = await handleSearchRequest(request, env, ctx);
				break;
			case '/':
				response = await serveR2Object(request, env, 'index.html');
				break;
			default:
				const objectKey = url.pathname.slice(1);
				response = await serveR2Object(request, env, objectKey);
				break;
		}

		return response;
	},
};

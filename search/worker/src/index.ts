import { Client } from '@neondatabase/serverless';
import { query } from './query';
import { validateSearchRequest } from './validation';

async function getClient(databaseUrl: string): Promise<Client> {
	if (databaseUrl === undefined) {
		throw new Error('DATABASE_URL is required');
	}

	const now = performance.now();
	const client = new Client(databaseUrl);
	await client.connect();
	console.log('Connected to database in', performance.now() - now, 'ms');
	return client;
}

async function fetchData(client: Client, queryParam: string, ctx: ExecutionContext): Promise<Response> {
	try {
		const start = performance.now();
		const results = await query(client, queryParam);
		const end = performance.now();
		console.log(`Query took ${end - start}ms`);
		ctx.waitUntil(client.end()); // Don't block on closing the connection
		return Response.json(results, {
			headers: {
				'Cache-Control': 'public, max-age=300', // Cache for 5 mins
			},
		});
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
	console.log('Querying for:', validation.queryParam);
	const response = await fetchData(client, validation.queryParam, ctx);
	return response;
}

async function serveR2Object(request: Request, env: Env, objectKey: string) {
	if (!objectKey) {
		return new Response('Not Found', { status: 404 });
	}

	console.log('Serving object:', objectKey);

	if (!env.BUCKET) {
		return new Response('Internal Server Error, bucket not found', { status: 500 });
	}

	const object = await env.BUCKET.get(objectKey);
	if (!object) {
		return new Response('Not Found', { status: 404 });
	}

	const response = new Response(object.body, {
		headers: {
			'Content-Type': object.httpMetadata!.contentType || 'application/octet-stream',
			'Cache-Control': 'public, max-age=3600', // Cache for 1 hour
		},
	});
	return response;
}

function applyCorsHeaders(response: Response) {
	response.headers.set('Access-Control-Allow-Origin', '*');
	response.headers.set('Access-Control-Allow-Methods', 'GET');
	return response;
}

export default {
	async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
		// This is a readonly api right now, we only support GET requests
		if (request.method !== 'GET') {
			return new Response('Method Not Allowed', { status: 405 });
		}

		const log = (message: string) => console.log(`[${request.method}]${url.pathname}${url.search} - ${message}`);

		const url = new URL(request.url);
		log('Request received');

		const cache = caches.default;
		let response = await cache.match(request);
		if (response) {
			log('Cache hit');
			return applyCorsHeaders(new Response(response.body, response));
		}

		switch (url.pathname) {
			case '/registry/docs/search':
			case '/search':
				response = await handleSearchRequest(request, env, ctx);
				break;
			case '/':
				response = await serveR2Object(request, env, 'index.html');
				break;
			default:
				if (url.pathname.startsWith('/registry/docs/')) {
					const objectKey = url.pathname.replace('/registry/docs/', '');
					response = await serveR2Object(request, env, objectKey);
					break;
				}
				const objectKey = url.pathname.slice(1);
				response = await serveR2Object(request, env, objectKey);
				break;
		}

		if (response.status === 200) {
			log('Cache miss, storing response');
			ctx.waitUntil(cache.put(request, response.clone()));
		}

		return applyCorsHeaders(response);
	},
};

import { Client, neon, neonConfig } from '@neondatabase/serverless';

import { DBClient } from './types';
import { getClient } from './client';
import { getTopProviders, query } from './query';
import { validateSearchRequest, validateTopProvidersRequest } from './validation';

async function fetchData(client: DBClient, queryParam: string, ctx: ExecutionContext): Promise<Response> {
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

async function fetchTopProviders(client: DBClient, limit: number, ctx: ExecutionContext): Promise<Response> {
	try {
		const start = performance.now();
		const results = await getTopProviders(client, limit);
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

	const client = await getClient(env.ENVIRONMENT, env.DATABASE_URL);
	console.log('Querying for:', validation.queryParam);
	const response = await fetchData(client, validation.queryParam, ctx);
	return response;
}

async function handleTopProvidersRequest(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
	const validation = validateTopProvidersRequest(request);
	if (validation.error) {
		return new Response(validation.error.message, { status: validation.error.status });
	}
	const client = await getClient(env.ENVIRONMENT, env.DATABASE_URL);
	console.log('Querying for top providers with limit:', validation.queryParam);
	const response = await fetchTopProviders(client, validation.queryParam, ctx);
	return response;
}

async function serveR2Object(request: Request, env: Env, objectKey: string) {
	if (!objectKey) {
		return new Response('Not Found', { status: 404 });
	}

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
			case '/providers/top':
				response = await handleTopProvidersRequest(request, env, ctx);
				break;
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

import { Client } from '@neondatabase/serverless';
import { query } from './query';

function validateRequest(request: Request): { queryParam: string } | Response {
	const url = new URL(request.url);

	// Validate the request method
	if (request.method !== 'GET') {
		return new Response('Method not allowed', { status: 405 });
	}

	// Validate the request path
	if (url.pathname !== '/') {
		return new Response('Not found', { status: 404 });
	}

	// Validate query parameters
	const queryParam = url.searchParams.get('q');
	if (!queryParam) {
		return new Response('No query provided', { status: 400 });
	}

	return { queryParam };
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

function getClient(databaseUrl: string): Client {
	if (databaseUrl === undefined) {
		throw new Error('DATABASE_URL is required');
	}

	const client = new Client(databaseUrl);
	client.connect();
	return client;
}

// Main fetch handler
export default {
	async fetch(request: Request, env: any, ctx: ExecutionContext): Promise<Response> {
		const client = getClient(env.DATABASE_URL);

		// Validate the incoming request
		const validation = validateRequest(request);
		if (validation instanceof Response) {
			return validation;
		}

		// Proceed with fetching data if validation passes
		return await fetchData(client, validation.queryParam, ctx);
	},
};

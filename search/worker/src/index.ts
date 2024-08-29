import { Client } from '@neondatabase/serverless';
import { query } from './query';
import { validateRequest } from './validation';

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

export default {
	async fetch(request: Request, env: any, ctx: ExecutionContext): Promise<Response> {
		const validation = validateRequest(request);
		if (validation.error) {
			return new Response(validation.error.message, { status: validation.error.status });
		}

		const client = await getClient(env.DATABASE_URL);
		return await fetchData(client, validation.queryParam, ctx);
	},
};

import { Client } from '@neondatabase/serverless';
import { DBClient } from "./types";
import { Client as PGClient } from "pg";

function getClientInstance(environment: string, databaseUrl: string): DBClient {
	if (environment == "dev") {
		return new PGClient(databaseUrl);
	} else {
		return new Client(databaseUrl);
	}
}

export async function getClient(environment: string, databaseUrl: string): Promise<DBClient> {
	if (databaseUrl === undefined) {
		throw new Error('DATABASE_URL is required');
	}

	const now = performance.now();
	const client = getClientInstance(environment, databaseUrl);
	await client.connect();
	console.log('Connected to database in', performance.now() - now, 'ms');
	return client;
}

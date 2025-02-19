import { Client } from '@neondatabase/serverless';
import { DBClient } from "./types";
import postgres from "postgres";

export class PGClient {
	db: postgres.Sql
	constructor(connection: string) {
		this.db = postgres(connection)
	}

	connect() {}
	end() : Promise<any> {
		return Promise.resolve();
	}

	query(query: string, queryParams: string[]): any {
		// Adapting to neonserverless/db return of rows
		return { rows: this.db.unsafe(query, queryParams)}
	}
}

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

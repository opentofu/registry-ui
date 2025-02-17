import { Client } from '@neondatabase/serverless';
import { Client as PGClient } from "pg";
import { DBClient } from "./types";

export function getClientType(environment: string, databaseUrl: string): DBClient {
	if (environment == "dev") {
		return new PGClient(databaseUrl);
	} else {
		return new Client(databaseUrl);
	}
}

import { Client } from '@neondatabase/serverless';
import { PGClient } from "./client";

export interface Entity {
	id: string;
	last_updated: Date;

	type: string;
	addr: string;
	version: string;
	title: string;
	description?: string;
	link_variables?: Record<string, any>;
}

export type DBClient = Client | PGClient;

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

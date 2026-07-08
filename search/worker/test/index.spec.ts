import { createExecutionContext, waitOnExecutionContext } from 'cloudflare:test';
import { env, exports } from 'cloudflare:workers';
import { describe, it, expect, vi } from 'vitest';
import worker from '../src/index';
import { query } from '../src/query';
import { Entity } from '../src/types';

const IncomingRequest = Request<unknown, IncomingRequestCfProperties>;

vi.mock('../src/query', () => {
	return {
		query: vi.fn().mockImplementation(() => {
			return [];
		}),
	};
});

vi.mock('@neondatabase/serverless', () => {
	return {
		Client: class {
			connect = vi.fn();
			end = vi.fn().mockResolvedValue(undefined);
			query = vi.fn().mockResolvedValue({ rows: [] });
		},
	};
});

describe('Validation', () => {
	it('responds 405 when not using GET', async () => {
		const request = new IncomingRequest('http://example.com', {
			method: 'POST',
		});

		const ctx = createExecutionContext();
		const response = await worker.fetch(request, env, ctx);

		await waitOnExecutionContext(ctx);

		expect(response.status).toBe(405);
	});

	it('responds 404 when requesting a path other than /', async () => {
		const request = new IncomingRequest('http://example.com/blah', {
			method: 'GET',
		});

		const ctx = createExecutionContext();
		const response = await worker.fetch(request, env, ctx);

		await waitOnExecutionContext(ctx);

		expect(response.status).toBe(404);
	});

	it('responds 400 when requesting /search without a query param', async () => {
		const request = new IncomingRequest('http://example.com/search', {
			method: 'GET',
		});

		const ctx = createExecutionContext();
		const response = await worker.fetch(request, env, ctx);

		await waitOnExecutionContext(ctx);

		expect(response.status).toBe(400);
	});
});

describe('Fetch handler happy path', () => {
	it('responds 200 when requesting /search with a query param', async () => {
		const request = new Request('http://example.com/search?q=hello', {
			method: 'GET',
		});

		const env = {
			ENVIRONMENT: 'production',
			DATABASE_URL: 'mocked-database-url',
		} as unknown as Env;

		const ctx = createExecutionContext();
		ctx.waitUntil = vi.fn();

		const response = await worker.fetch(request, env, ctx);

		expect(response.status).toBe(200);
		expect(await response.json()).toEqual([]);
	});
});

describe('End-to-end', () => {
	it('serves an object from R2 with cache and CORS headers', async () => {
		// Serving r2 files
		await env.BUCKET.put('example/index.html', '<html>hello registry</html>', {
			httpMetadata: { contentType: 'text/html' },
		});

		const response = await exports.default.fetch('https://example.com/registry/docs/example/index.html');

		expect(response.status).toBe(200);
		expect(response.headers.get('Content-Type')).toBe('text/html');
		expect(response.headers.get('Cache-Control')).toBe('public, max-age=3600');
		expect(response.headers.get('Access-Control-Allow-Origin')).toBe('*');
		expect(await response.text()).toBe('<html>hello registry</html>');
	});

	it('runs a search and returns the matching result with cache and CORS headers', async () => {
		// Searching. using a mocked response, just to ensure that we're setting cache headers correctly.
		// we do not want to test the db here.
		const result: Entity = {
			id: 'hashicorp/aws',
			last_updated: new Date('2024-01-01T00:00:00.000Z'),
			type: 'provider',
			addr: 'hashicorp/aws',
			version: 'v5.0.0',
			title: 'aws',
			description: 'The AWS provider',
		};
		vi.mocked(query).mockResolvedValueOnce([result]);

		const response = await exports.default.fetch('https://example.com/search?q=aws');

		expect(response.status).toBe(200);
		expect(response.headers.get('Content-Type')).toContain('application/json');
		expect(response.headers.get('Cache-Control')).toBe('public, max-age=300');
		expect(response.headers.get('Access-Control-Allow-Origin')).toBe('*');
		expect(await response.json()).toEqual([{ ...result, last_updated: '2024-01-01T00:00:00.000Z' }]);
	});
});

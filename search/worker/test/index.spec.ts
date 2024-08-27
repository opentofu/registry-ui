import { env, createExecutionContext, waitOnExecutionContext } from 'cloudflare:test';
import { describe, it, expect, vi } from 'vitest';
import worker from '../src/index';

const IncomingRequest = Request<unknown, IncomingRequestCfProperties>;

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

	it('responds 400 when requesting / without a query param', async () => {
		const request = new IncomingRequest('http://example.com/', {
			method: 'GET',
		});

		const ctx = createExecutionContext();
		const response = await worker.fetch(request, env, ctx);

		await waitOnExecutionContext(ctx);

		expect(response.status).toBe(400);
	});
});

describe('Fetch handler happy path', () => {
	it('responds 200 when requesting / with a query param', async () => {
		// Move the mock inside the test
		vi.mock('../src/query', () => {
			return {
				query: vi.fn().mockImplementation(() => {
					return [];
				}),
			};
		});

		vi.mock('@neondatabase/serverless', () => {
			const mockConnect = vi.fn();
			const mockEnd = vi.fn();
			return {
				Client: vi.fn().mockImplementation(() => ({
					connect: mockConnect,
					end: mockEnd,
				})),
			};
		});

		const request = new Request('http://example.com/?q=hello', {
			method: 'GET',
		});

		const env = {
			DATABASE_URL: 'mocked-database-url',
		};

		const ctx = createExecutionContext();
		ctx.waitUntil = vi.fn();

		const response = await worker.fetch(request, env, ctx);

		expect(response.status).toBe(200);
		expect(await response.json()).toEqual([]);
	});
});

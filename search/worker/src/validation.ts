interface ValidationSuccess<T> {
	queryParam: T;
	error?: never; // Ensures this type does not have an error
}

interface ValidationError {
	queryParam?: never; // Ensures this type does not have a queryParam
	error: {
		message: string;
		status: number;
	};
}

type ValidationResul<T> = ValidationSuccess<T> | ValidationError;

const newValidationErrorResult = (message: string, status: number): ValidationError => ({ error: { message, status } });

export function validateSearchRequest(request: Request): ValidationResult<string> {
	const url = new URL(request.url);

	if (request.method !== 'GET') {
		return newValidationErrorResult('Method not allowed', 405);
	}

	const queryParam = url.searchParams.get('q');
	if (!queryParam) {
		return newValidationErrorResult('Query parameter "q" is required', 400);
	}

	return { queryParam };
}

export function validateTopProvidersRequest(request: Request): ValidationResult<number> {
	const url = new URL(request.url);

	if (request.method !== 'GET') {
		return newValidationErrorResult('Method not allowed', 405);
	}

	const limitParam = url.searchParams.get('limit');
	if (!limitParam) {
		return newValidationErrorResult('Query parameter "limit" is required', 400);
	}

	const limit = parseInt(limitParam, 10);
	if (isNaN(limit) || limit <= 0) {
		return newValidationErrorResult('Query parameter "limit" must be a positive integer', 400);
	}

	// don't allow limit to be greater than 500
	if (limit > 500) {
		return newValidationErrorResult('Query parameter "limit" must be less than or equal to 500', 400);
	}

	return { queryParam: limit };
}

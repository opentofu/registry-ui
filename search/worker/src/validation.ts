interface ValidationSuccess {
	queryParam: string;
	error?: never; // Ensures this type does not have an error
}

interface ValidationError {
	queryParam?: never; // Ensures this type does not have a queryParam
	error: {
		message: string;
		status: number;
	};
}

type ValidationResult = ValidationSuccess | ValidationError;

const newValidationErrorResult = (message: string, status: number): ValidationError => ({ error: { message, status } });

export function validateRequest(request: Request): ValidationResult {
	const url = new URL(request.url);

	if (request.method !== 'GET') {
		return newValidationErrorResult('Method not allowed', 405);
	}

	if (url.pathname !== '/') {
		return newValidationErrorResult('Not found', 404);
	}

	const queryParam = url.searchParams.get('q');
	if (!queryParam) {
		return newValidationErrorResult('Query parameter "q" is required', 400);
	}

	return { queryParam };
}

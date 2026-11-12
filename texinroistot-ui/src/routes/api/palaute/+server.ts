import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';
import { proxiedResponse } from '$lib/server/proxy-auth';

// Feedback from the Info page. The sender's address goes along for the
// backend's rate limit.
export const POST: RequestHandler = async ({ request, fetch, getClientAddress }) => {
	const response = await fetch(`${getBackendHost()}/api/feedback`, {
		method: 'POST',
		headers: {
			'content-type': 'application/json',
			'x-forwarded-for': getClientAddress()
		},
		body: await request.text()
	});

	return proxiedResponse(response);
};

import type { RequestHandler } from './$types';
import { getBackendHost } from '$lib/server/backend-host';
import { authProxyHeaders, proxiedResponse } from '$lib/server/proxy-auth';

export const POST: RequestHandler = async ({ request, fetch }) => {
	const body = new URLSearchParams();
	const contentType = request.headers.get('content-type') ?? '';

	if (contentType.includes('application/json')) {
		const payload = (await request.json().catch(() => null)) as {
			credential?: string;
			g_csrf_token?: string;
		} | null;

		if (payload?.credential) {
			body.append('credential', payload.credential);
		}
		if (payload?.g_csrf_token) {
			body.append('g_csrf_token', payload.g_csrf_token);
		}
	} else {
		const formData = await request.formData();
		for (const [key, value] of formData.entries()) {
			if (typeof value === 'string') {
				body.append(key, value);
			}
		}
	}

	if (!body.get('credential')) {
		return new Response(JSON.stringify({ error: 'Google credential missing' }), {
			status: 400,
			headers: { 'content-type': 'application/json' }
		});
	}

	const headers = authProxyHeaders(request, {
		'content-type': 'application/x-www-form-urlencoded'
	});

	const response = await fetch(`${getBackendHost()}/api/login`, {
		method: 'POST',
		headers,
		body: body.toString(),
		redirect: 'manual'
	});

	return proxiedResponse(response);
};

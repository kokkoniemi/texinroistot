const defaultOptions: RequestInit = {
    method: 'GET',
    headers: {
        Accept: "application/json",
    },
    mode: 'same-origin',
    referrer: 'same-origin',

}

const httpClient = {
    get: async (url: string, params?: Record<string, string>) =>
        httpClient.request(serializeUrl(url, params), { method: 'GET' }),
    post: async (url: string, params?: Record<string, string>, body?: BodyInit) =>
        httpClient.request(serializeUrl(url, params), {method: 'POST', body}),
    request: async (url: string, config: RequestInit | undefined) => {
        const res = await fetch(url, {
            ...defaultOptions,
            ...config
        })
        return res.json()
    }
}

const serializeUrl = (url: string, params?: Record<string, string>) => {
    let res = url
    if (params) {
        res = res + '?' + new URLSearchParams(params).toString()
    }
    return res
}

const auth = {
    me: {
        read: () => httpClient.get("/api/me"),
    },
    logout: {
        post: () => httpClient.post("/api/logout")
    }
}

const admin = {
    users: {
        list: () => httpClient.get("/api/admin/users"),
    },
}

export const api = {auth, admin}

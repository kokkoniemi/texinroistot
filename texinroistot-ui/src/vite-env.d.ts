/// <reference types="svelte" />
/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly VITE_GOOGLE_OAUTH2_CLIENT_ID: string
    readonly VITE_OAUTH_LOGIN_URL: string
}
  
interface ImportMeta {
    readonly env: ImportMetaEnv
}
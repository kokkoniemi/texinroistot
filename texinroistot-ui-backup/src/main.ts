import App from './App.svelte'
import { appStatus } from './lib/store';

declare var window: {
  google: any
  onload: (fn: any) => void
}

window.onload = () => {
  window.google.accounts.id.initialize({
    client_id: import.meta.env.VITE_GOOGLE_OAUTH2_CLIENT_ID,
    context: "signin",
    ux_mode: 'redirect',
    login_uri: import.meta.env.VITE_OAUTH_LOGIN_URL,
    nonce: '',
    auto_select: false
  });

  appStatus.update((val) => Object.assign(val, { loginInitialized: true }))
}
const app = new App({
  target: document.getElementById('app') as HTMLElement,
})

export default app

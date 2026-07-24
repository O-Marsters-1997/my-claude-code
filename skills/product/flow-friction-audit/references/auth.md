# Auth precondition — Plain support-app

The `support-app` authenticates with **WorkOS AuthKit** (`@workos-inc/authkit-js`) using **Sign in with Google**. Sign-in redirects through WorkOS' hosted flow to Google and back to `/callback/`; the session then lives in the Redux store and pages are gated by the `withAuthenticationRequired` HOC.

**Recognising the logged-out state.** An unauthenticated page renders a full-screen loading overlay (the HOC's `FullscreenLoadingOverlay`) or the WorkOS/Google sign-in screen — never the app content. If a `snapshot` shows either, treat the session as unauthenticated.

**What to do.** Do not drive the Google sign-in — Google blocks automated logins. Pause and ask the user to complete "Sign in with Google" manually in the browser session you already opened, then continue the walk once app content renders. Because the audit keeps one browser session for the whole run, a single manual login at the start carries through every step.

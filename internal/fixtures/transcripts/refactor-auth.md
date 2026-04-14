@@STATE: running
@@DELAY: 300
## refactor auth module to use OAuth2 PKCE

@@DELAY: 400
Reading the repo structure to find the current auth implementation…

@@DELAY: 600
▼ grep  · 82ms · exit 0
```
src/auth/oauth.go:12:func exchangeCode(...)
src/auth/oauth.go:48:func refresh(...)
src/auth/session.go:9:type Session struct {...}
```

@@DELAY: 500
Found `src/auth/oauth.go`. It uses the classic auth-code flow with a shared client secret — we need to move to PKCE so the native client never holds the secret.

@@DELAY: 700
Let me patch `exchangeCode` to accept a `codeVerifier` and drop the `clientSecret` argument.

@@DELAY: 900
▼ edit  · src/auth/oauth.go · exit 0
```diff
-func exchangeCode(ctx context.Context, code, clientSecret string) (*Token, error) {
-    form := url.Values{"code": {code}, "client_secret": {clientSecret}}
+func exchangeCode(ctx context.Context, code, codeVerifier string) (*Token, error) {
+    form := url.Values{"code": {code}, "code_verifier": {codeVerifier}}
     form.Set("grant_type", "authorization_code")
```

@@DELAY: 500
Now updating the caller in `cmd/login/main.go` to generate a verifier and challenge pair.

@@DELAY: 700
Running the test suite to confirm nothing regressed…

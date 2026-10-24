package main

import (
	"fmt"
	"html"
)

func createFormPage(appPath string, manifestJson []byte, stateSecret string) []byte {
	return fmt.Appendf(nil,
		`<form id="create-app" action="https://github.com%s" method="post">`+
			`<input type="hidden" name="manifest" value="%s">`+
			`<input type="hidden" name="state" value="%s">`+
			`</form><script>document.getElementById('create-app').submit()</script>`,
		html.EscapeString(appPath),
		html.EscapeString(string(manifestJson)),
		html.EscapeString(stateSecret),
	)
}

func successPage(privateKeyPath string) []byte {
	return fmt.Appendf(nil, `
		<!doctype html>
		<html>

		<head>
			<title>GitHub App Created</title>
		</head>

		<body style="font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #343434; height: 100vh; overflow: hidden;">
			<div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%%;">
				<div
					style="display: flex; flex-direction: column; gap: 1em; background-color: #90c1cf; padding: 20px; border-radius: 8px; max-width: 600px">
					<div>
						<h1>GitHub App Created Successfully!</h1>
						<p>The App ID is now available in your env, and the private key was saved to
							<code style="font-size: 12pt; font-weight: bold">%s</code>.
						</p>
					</div>
					<span>You may now close this window.</span>
				</div>
			</div>
		</body>

		</html>
		`, privateKeyPath)
}

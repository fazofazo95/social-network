Frontend test for signup

How to use

1. Start the backend server (from `backend/`):

   - Build and run with Go (requires Go installed):

     go run server.go

   The server will listen on http://localhost:8080.

2. Serve the frontend folder as static files. Two quick options:

   - Using Python (if installed):

     python -m http.server 3000

     Then open http://localhost:3000 in your browser.

   - Or just open `frontend/index.html` with a Live Server extension or similar. Note: opening the file via file:// may trigger CORS/preflight issues; using a local static server is recommended.

3. Fill the form and click "Sign up". The page will display the backend response.

Notes

- For local testing a permissive CORS header is set on the signup handler. Tighten or remove it for production.
- If you already have users in the DB, duplicate emails or usernames will return HTTP 409 with an explanatory message.

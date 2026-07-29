/* ==========================================================================
   Marginalia — API configuration

   This matches the ACTUAL Go/Echo backend in this repo (not a guessed
   contract). Key differences from a "conventional" REST API that the
   original frontend assumed:

   REAL ENDPOINTS
   --------------
   POST   {BASE}/signup              { username, password } -> 202 { user: <id>, token }
   POST   {BASE}/login               { username, password } -> 202 { user: <id>, token }
   GET    {BASE}/weblog                                     -> 200 [ Board ]
   GET    {BASE}/weblog/:id                                 -> 200 Board | 403 | 404
   POST   {BASE}/weblog        JSON  { title, content, is_private, img_path } -> 201 Board
   DELETE {BASE}/weblog/:id                                 -> 204
   POST   {BASE}/weblog/:id/share    { usernames: [string] } -> 200
   GET    {BASE}/weblog/:id/comment                          -> 200 [ Comment ]
   POST   {BASE}/weblog/:id/comment  { content }              -> 201 Comment
   DELETE {BASE}/weblog/:id/comment/:commentId                -> 204

   Board = { id, author_id, title, content, is_private, img_path }
   Comment = { id, author_id, board_id, content }

   THINGS THE BACKEND DOES NOT PROVIDE (so the UI can't show them):
   - No created_at/timestamps on boards or comments.
   - No username on boards/comments — only a numeric author_id. There is
     no "look up username by id" endpoint, so people are shown as
     "User #<id>" rather than "@handle" everywhere except the person
     currently logged in (whose own username we already know locally).
   - No image upload — img_path is just a plain text field (a path or
     URL you type in), not a file upload.
   - No GET /me and no /logout route — auth is a bare bearer token with
     no session-check endpoint, so the frontend persists the token and
     the user's own {id, username} in localStorage after login/signup
     and trusts it until a request comes back 401.
   - The share endpoint is separate from board creation, so creating a
     private board with people to share it with takes two API calls.
   ========================================================================== */

window.MARGINALIA_CONFIG = {
  // The Go server in this repo listens on :8080 with no path prefix and
  // isn't served from the same origin as this static frontend (docker-compose
  // here only stands up Postgres). Point this at wherever `go run .` /
  // the built binary is actually listening.
  API_BASE: 'https://webloggoraz-production.up.railway.app',
};

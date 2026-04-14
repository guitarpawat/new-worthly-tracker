package frontendassets

import "embed"

// Files contains the static Wails frontend used by the MVP home page.
//
//go:embed index.html app.js styles.css appicon.png js/*.js styles/*.css vendor/*.js
var Files embed.FS

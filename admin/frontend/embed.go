package frontend

import "embed"

// Assets는 별도의 Node 런타임 없이 Admin 바이너리에 포함된다.
//go:embed index.html app.js forms.js app.css version.js
var Assets embed.FS

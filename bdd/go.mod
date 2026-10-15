module github.com/larsartmann/go-graph-rag/bdd

go 1.27

// The replace keeps local and CI builds pinned to the sibling core checkout;
// it is inert for consumers (replace directives only apply in the main
// module, so proxy resolution uses the required version above). Keep the
// require in sync with the newest released core tag when cutting bdd tags.
replace github.com/larsartmann/go-graph-rag => ../

require (
	github.com/larsartmann/go-graph-rag v0.0.0-00010101000000-000000000000
	github.com/onsi/gomega v1.44.0
)

require (
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.1.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/samber/lo v1.53.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	modernc.org/libc v1.78.0 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	modernc.org/sqlite v1.60.1 // indirect
)

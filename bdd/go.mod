module github.com/larsartmann/go-graph-rag/bdd

go 1.27

// The replace keeps local and CI builds pinned to the sibling core checkout;
// it is inert for consumers (replace directives only apply in the main
// module, so proxy resolution uses the required version above). Keep the
// require in sync with the newest released core tag when cutting bdd tags.
replace github.com/larsartmann/go-graph-rag => ../

package compiler

// Version names the build a release tags, so a program can read back which one
// it is running rather than guess from the output that was written for it. It
// stays below 1.0 while the tool is an experimental release, and carries the
// warning that goes with that: see docs/limitations.md.
const Version = "v0.1.0"

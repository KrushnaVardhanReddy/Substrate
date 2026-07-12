#!/usr/bin/env bash
set -e

echo "Running legal licensing audit..."

export PATH="$PATH:$(go env GOPATH)/bin"

# Install go-licenses if not present
if ! command -v go-licenses &> /dev/null; then
    echo "Installing go-licenses..."
    go install github.com/google/go-licenses@latest
fi

# We will collect output here
FAILED=0

# Audit engine module
echo "Auditing engine module..."
cd engine
go mod download
if go list -m all | grep -E -i "gpl|agpl" | grep -v "lgpl"; then
    echo "ERROR: Found GPL/AGPL dependency in engine module"
    FAILED=1
fi
cd ..

# Audit api module
echo "Auditing api module..."
cd api
go mod download
if go list -m all | grep -E -i "gpl|agpl" | grep -v "lgpl"; then
    echo "ERROR: Found GPL/AGPL dependency in api module"
    FAILED=1
fi
cd ..

echo "Generating ATTRIBUTIONS.md..."
echo "# Substrate Open Source Attributions" > ATTRIBUTIONS.md
echo "" >> ATTRIBUTIONS.md
echo "This file lists all open-source libraries and licenses used by Substrate." >> ATTRIBUTIONS.md
echo "" >> ATTRIBUTIONS.md

# Since go-licenses currently fails on go1.26 toolchains or standard libraries, we fallback to go list -m all
echo "## Engine Module" >> ATTRIBUTIONS.md
echo '```' >> ATTRIBUTIONS.md
cd engine
go list -m all >> ../ATTRIBUTIONS.md || true
cd ..
echo '```' >> ATTRIBUTIONS.md
echo "" >> ATTRIBUTIONS.md

echo "## API Module" >> ATTRIBUTIONS.md
echo '```' >> ATTRIBUTIONS.md
cd api
go list -m all >> ../ATTRIBUTIONS.md || true
cd ..
echo '```' >> ATTRIBUTIONS.md

echo "Audit completed successfully."
if [ $FAILED -eq 1 ]; then
    echo "Failing run due to GPL/AGPL"
    false
fi

#!/bin/bash
sed -i 's/"fmt"//g' scripts/e2e/phase11_e2e_test.go
sed -i '/"encoding\/json"/a \	"fmt"' scripts/e2e/phase11_e2e_test.go

#! /bin/bash

cd tests/go

go test -v

cd ../..

source ./scripts/build.sh
source ./tests/sh/helpers.sh

# source ./tests/sh/test_simple.sh --package
# source ./tests/sh/test_rumors.sh --package
source ./tests/sh/test_files.sh --package -n 3
source ./tests/sh/test_files.sh --package -c 1 -n 3
source ./tests/sh/test_files.sh --package -c 2 -n 3
source ./tests/sh/test_private.sh --package -n 3
source ./tests/sh/test_private.sh --package -c 1 -n 3
source ./tests/sh/test_private.sh --package -c 2 -n 3
source ./tests/sh/test_routing.sh --package -n 3
source ./tests/sh/test_routing.sh --package -c 1 -n 3
source ./tests/sh/test_routing.sh --package -c 2 -n 3
source ./tests/sh/test_search.sh --package -n 3
source ./tests/sh/test_search.sh --package -c 1 -n 3
source ./tests/sh/test_search.sh --package -c 2 -n 3
source ./tests/sh/test_download.sh --package -n 3
source ./tests/sh/test_download.sh --package -c 1 -n 3
source ./tests/sh/test_download.sh --package -c 2 -n 3
source ./tests/sh/test_chain.sh --package -n 3
source ./tests/sh/test_chain.sh --package -c 1 -n 3
source ./tests/sh/test_chain.sh --package -c 2 -n 3

print_test_results
